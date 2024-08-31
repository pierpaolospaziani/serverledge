package scheduling

import (
	"log"
	"sort"

    // tg "github.com/galeone/tfgo"
	tf "github.com/galeone/tensorflow/tensorflow/go"

	"github.com/grussorusso/serverledge/internal/node"
)

type decisionEngineDQN struct {
	mg *metricGrabberDQN
}

type Model struct {
    Session *tf.Session
    Graph   *tf.Graph
}

var dqnModel *Model

type State struct {
    PercAvailableLocalMemory float32
    CanExecuteOnEdge         float32
    FunctionId               []float32
    ClassId                  []float32
    HasBeenOffloaded         float32 	// == !CanDoOffloading (do not remove the ! cause the NN has been trained with has_been_offloaded)
}


func LoadModel(modelPath string) *Model {
    dqnModel, err := tf.LoadSavedModel(modelPath, []string{"serve"}, nil)
    if err != nil {
        return nil
    }
    return &Model{
        Session: dqnModel.Session,
        Graph:   dqnModel.Graph,
    }
    return nil
}


func (m *Model) Predict(s State, actionFilter []bool) (int, error) {
    state := []float32{
        s.PercAvailableLocalMemory,
        s.CanExecuteOnEdge,
    }
    state = append(state, s.FunctionId...)
    state = append(state, s.ClassId...)
    state = append(state, s.HasBeenOffloaded)

    inputTensor, err := tf.NewTensor([][]float32{state})
    if err != nil {
        return 0, err
    }

    /*
    	saved_model_cli show --dir tf_model --all 
    	saved_model_cli show --dir tf_model --tag_set serve --signature_def serving_default
    */
    /* MODEL */
    result, err := m.Session.Run(
        map[tf.Output]*tf.Tensor{
            m.Graph.Operation("serving_default_keras_tensor").Output(0): inputTensor,
        },
        []tf.Output{
            m.Graph.Operation("StatefulPartitionedCall_1").Output(0),
        },
        nil,
    )
    /* TF_MODEL */
    // result, err := m.Session.Run(
    //     map[tf.Output]*tf.Tensor{
    //         m.Graph.Operation("serving_default_inputs").Output(0): inputTensor,
    //     },
    //     []tf.Output{
    //         m.Graph.Operation("StatefulPartitionedCall").Output(0),
    //     },
    //     nil,
    // )
    if err != nil {
        return 0, err
    }

    // create a slice with predictions
    prediction := result[0].Value().([][]float32)[0]

    // filter the actions
    for i, allowed := range actionFilter {
        if !allowed {
            prediction[i] = 0.0
        }
    }

    // return the index of highest value
    action := 0
    maxValue := float32(-1)
    for i, value := range prediction {
        if value > maxValue {
            action = i
            maxValue = value
        }
    }
    return action, nil
    // return 3, nil
}


func oneHotEncoding(list []string, str string) []float32 {
	indexMap := make(map[string]int)
	for i, v := range list {
		indexMap[v] = i
	}
	oneHot := make([]float32, len(list))
	if idx, exists := indexMap[str]; exists {
		oneHot[idx] = 1.0
	}
	return oneHot
}


func getState(r *scheduledRequest) State {
	percAvailableLocalMemory := float32(node.Resources.AvailableMemMB + node.FreeableMemory(r.Fun)) / float32(node.Resources.MaxMemMB)
	// log.Printf("percAvailableLocalMemory = %f", percAvailableLocalMemory)

	canExecuteOnEdge := float32(1.0)
	url := pickEdgeNodeForOffloading(r)
	if url == "" {
		canExecuteOnEdge = 0.0
	}
	// log.Printf("canExecuteOnEdge = %t", canExecuteOnEdge)

	functions, err := r.Fun.GetAll()
	if err != nil {
		message := " ### ERRORE (decisionEngineDQN): Fun.GetAll()!"
		log.Printf(message)
		panic(err)
	}
	sort.Strings(functions)	// need to sort cause Go mixes maps and NN needs functionId in order
	functionId := oneHotEncoding(functions, r.Fun.Name)
	// log.Printf("functionId = %v -> %v", functions, functionId)

	classList := make([]string, 0, len(Classes))
    for key := range Classes {
        classList = append(classList, key)
    }
	sort.Strings(classList)	// need to sort cause Go mixes maps and NN needs classId in order
	classId := oneHotEncoding(classList, r.ClassService.Name)
	// log.Printf("classId = %v -> %v", classList, classId)

	state := State{
        PercAvailableLocalMemory: percAvailableLocalMemory,
        CanExecuteOnEdge:         canExecuteOnEdge,
        FunctionId:               functionId,
        ClassId:                  classId,
        HasBeenOffloaded:         0.0,
    }
    if !r.CanDoOffloading {
    	state.HasBeenOffloaded = 1.0
    }
    log.Printf("State = %+v", state)
	return state
}


func actionFilter(state State, r *scheduledRequest) []bool {
	actionFilter := []bool{true, true, true, true}
	availableMemory := float32(node.Resources.MaxMemMB) * state.PercAvailableLocalMemory
	canExecuteLocally := canExecute(r.Fun)
	log.Println("availableMemory = ",availableMemory, "| canExecuteLocally = canExecuteLocally")
	canExecuteOnCloud := state.HasBeenOffloaded == 0.0
	canExecuteOnEdge := state.CanExecuteOnEdge == 1.0 && state.HasBeenOffloaded == 0.0
	if !canExecuteLocally {
		actionFilter[0] = false
	}
	if !canExecuteOnCloud {
		actionFilter[1] = false
	}
	if !canExecuteOnEdge {
		actionFilter[2] = false
	}
	return actionFilter
}


func (d *decisionEngineDQN) Decide(r *scheduledRequest) int {

	/*
		CONTROLLARE SE E' CORRETTO FreeableMemory
	*/
	// retrieve state and the possible best edge node to offload to
	state := getState(r)

	// boolean list of allowed actions
	actionFilter := actionFilter(state, r)

	// check how many actions can be taken
	numActionsAllowed := 0
	action := 0
    for i, value := range actionFilter {
        if value {
            numActionsAllowed++
            action = i
        }
    }
    // if there is more than 1 let the model choose
    if numActionsAllowed > 1 {
    	var err error
    	action, err = dqnModel.Predict(state, actionFilter)
	    if err != nil {
	        log.Println("Error predicting:", err)
	        return -1
	    }
    }

	log.Println("Filter:",actionFilter,"-> Action =", action)

    // map simulator action to Serverledge
    //  - simulator:   LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
    //  - Serverledge: DROP(0)-LOCAL(1)-CLOUD(2)-EDGE(3)
    action = (action + 1) % 4

	if action == DROP_REQUEST {
		d.mg.addStats(r,true)
	}
	return action
}


func (d *decisionEngineDQN) InitDecisionEngine() {
	// model initialization
    modelPath := "model"
    dqnModel = LoadModel(modelPath)
    if dqnModel == nil {
        log.Println("Error loading model")
        return
    }
    d.mg = InitMG()
}


// VEDERE SE SERVE, IL MODELLO VA CHIUSO SOLO QUANDO SPEGNI TUTTO, MA DOVE?
func (d *decisionEngineDQN) CloseSession() {
	dqnModel.Session.Close()
}


func (d *decisionEngineDQN) Completed(r *scheduledRequest, offloaded int) {
	// log.Println("COMPLETED: in decisionEngineDQN")
	d.mg.addStats(r,false)
}


func (d *decisionEngineDQN) GetGrabber() metricGrabber {
	// VEDERE COSA DEVO FARCI
	return nil
}
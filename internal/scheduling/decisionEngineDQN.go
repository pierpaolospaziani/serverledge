package scheduling

import (
	"log"
	"sort"

    // tg "github.com/galeone/tfgo"
	tf "github.com/galeone/tensorflow/tensorflow/go"

	"github.com/grussorusso/serverledge/internal/node"
	"github.com/grussorusso/serverledge/internal/config"

	"encoding/json"
    "fmt"
    "os"
    "sync"
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

var stateMutex sync.Mutex


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

    // log.Println("[DE_DQN] Predictions-pre: ", prediction)

    // filter the actions
    for i, allowed := range actionFilter {
        if !allowed {
            prediction[i] = 0.0
        }
    }

    // log.Println("[DE_DQN] Predictions-post:", prediction)

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
	log.Println("\n[DE_DQN]",r.Fun.Name, r.ClassService.Name)
	log.Println("[DE_DQN]",node.WarmStatus())
	percAvailableLocalMemory := float32(node.Resources.MaxMemMB - node.Resources.BusyMemMB) / float32(node.Resources.MaxMemMB)
	log.Printf("[DE_DQN] AvailableMemMB = %f", float32(node.Resources.AvailableMemMB))
	log.Printf("[DE_DQN] BusyMemMB = %f", float32(node.Resources.BusyMemMB))
	log.Printf("[DE_DQN] MaxMemMB = %f", float32(node.Resources.MaxMemMB))
	log.Printf("[DE_DQN] WarmMemory = %f", float32(node.CountWarmMemory()))
	if node.Resources.MaxMemMB != node.Resources.AvailableMemMB + node.Resources.BusyMemMB + node.CountWarmMemory(){
		panic("IL CONTO NON TORNA!")
	}
	log.Printf("[DE_DQN] percAvailableLocalMemory = %f", percAvailableLocalMemory)

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
	// log.Printf("[DE_DQN] functionId = %v -> %v", functions, functionId)

	classList := make([]string, 0, len(Classes))
    for key := range Classes {
        classList = append(classList, key)
    }
	sort.Strings(classList)	// need to sort cause Go mixes maps and NN needs classId in order
	classId := oneHotEncoding(classList, r.ClassService.Name)
	log.Printf("[DE_DQN] classId = %v -> %v", classList, classId)

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
    log.Printf("[DE_DQN] State = %+v", state)
	return state
}


func actionFilter(state State, r *scheduledRequest) []bool {
	actionFilter := []bool{true, true, true, true}
	// availableMemory := float32(node.Resources.MaxMemMB) * state.PercAvailableLocalMemory
	canExecuteLocally := canExecute(r.Fun)
	// log.Println("[DE_DQN] availableMemory =",availableMemory, "| canExecuteLocally =",canExecuteLocally)
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
	stateMutex.Lock()
	defer stateMutex.Unlock()
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
	        log.Println("[DE_DQN] Error predicting:", err)
	        return -1
	    }
    }

	log.Println("[DE_DQN] Filter:",actionFilter,"-> Action =", action)
	// log.Println("[DE_DQN] Action =", action)

	// ------------------------------------------------------------------------
    filePath := "state_action.json"

	tuple := StateActionTuple{
		MaxMemMB:		float32(node.Resources.MaxMemMB),
        AvailableMemMB: float32(node.Resources.AvailableMemMB),
        BusyMemMB: 		float32(node.Resources.BusyMemMB),
        Perc: 			float32(node.Resources.MaxMemMB - node.Resources.BusyMemMB) / float32(node.Resources.MaxMemMB),
        State:        	state,
        ActionFilter: 	actionFilter,
        Action:       	action,
    }

	if err := saveStateActionToFile(tuple, filePath, &mutex); err != nil {
        fmt.Println("Errore nel salvare la tupla:", err)
    }
	// ------------------------------------------------------------------------

    // map simulator action to Serverledge
    //  - simulator:   LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
    //  - Serverledge: DROP(0)-LOCAL(1)-CLOUD(2)-EDGE(3)
    action = (action + 1) % 4

	if action == DROP_REQUEST {
		d.mg.addStats(r,true,false)
	}
	return action
}


// ------------------------------------------------------------------------

type StateActionTuple struct {
	MaxMemMB		float32
	AvailableMemMB	float32
	BusyMemMB		float32
	Perc 			float32
    State        	State
    ActionFilter 	[]bool
    Action       	int
}

var mutex sync.Mutex

func saveStateActionToFile(tuple StateActionTuple, filePath string, mutex *sync.Mutex) error {
    // Converti la struttura in JSON
    jsonData, err := json.Marshal(tuple)
    if err != nil {
        return err
    }
    // Aggiungi una nuova linea al JSON per separare le tuple
    jsonData = append(jsonData, '\n')
    // Blocca il mutex per evitare condizioni di corsa
    mutex.Lock()
    defer mutex.Unlock()
    // Apri il file in modalità append
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    // Scrivi i dati JSON nel file
    if _, err := file.Write(jsonData); err != nil {
        return err
    }
    return nil
}
// ------------------------------------------------------------------------


func (d *decisionEngineDQN) InitDecisionEngine() {
	// model initialization
	modelPath := config.GetString(config.DQN_MODEL_PATH, "dqn_models/model")
    dqnModel = LoadModel(modelPath)
    if dqnModel == nil {
        log.Println("[DE_DQN] Error loading model")
        return
    }
    d.mg = InitMG()
}


// VEDERE SE SERVE, IL MODELLO VA CHIUSO SOLO QUANDO SPEGNI TUTTO, MA DOVE?
func (d *decisionEngineDQN) CloseSession() {
	dqnModel.Session.Close()
}


func (d *decisionEngineDQN) Completed(r *scheduledRequest, offloaded int) {
	// log.Println("[DE_DQN] COMPLETED: in decisionEngineDQN")
	offloadDrop := offloaded != 0
	d.mg.addStats(r,false,offloadDrop)
}


func (d *decisionEngineDQN) GetGrabber() metricGrabber {
	// VEDERE COSA DEVO FARCI
	return nil
}
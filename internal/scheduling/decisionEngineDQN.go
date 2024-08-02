package scheduling

import (
	"log"

    "github.com/tensorflow/tensorflow/tensorflow/go"
    "github.com/tensorflow/tensorflow/tensorflow/go/saved_model"

	"github.com/grussorusso/serverledge/internal/node"
)

type decisionEngineDQN struct {
}

/*
type Model struct {
    Session *tensorflow.Session
    Graph   *tensorflow.Graph
}

func NewModel(modelPath string) (*Model, error) {
    model, err := tensorflow.LoadSavedModel(modelPath, []string{"serve"}, nil)
    if err != nil {
        return nil, fmt.Errorf("error loading model: %v", err)
    }
    return &Model{
        Session: model.Session,
        Graph:   model.Graph,
    }, nil
}
*/

type Model struct {
    model *saved_model.SavedModel
}

func LoadModel(modelPath string) (*Model, error) {
	// the 'serve' tag indicates that the model is intended to be used for inference or serving
    model, err := saved_model.Load(modelPath, []string{"serve"}, nil)
    if err != nil {
        return nil, err
    }
    return &Model{model: model}, nil
}

func (m *Model) Predict(state []float32, actionFilter []bool) ([]float32, error) {
	/*
    	VEDI QUALE DEI 2 VA BENE
    */
    // inputTensor, err := tensorflow.NewTensor(inputData)
    inputTensor, err := tensorflow.NewTensor([1][len(state)]float32{state})
    if err != nil {
        return nil, err
    }

    /*
    	saved_model_cli show --dir tf_model --all 
    	saved_model_cli show --dir tf_model --tag_set serve --signature_def serving_default
    */
    result, err := m.model.Session.Run(
        map[tensorflow.Output]*tensorflow.Tensor{
            m.model.Graph.Operation("serving_default_input").Output(0): inputTensor,
        },
        []tensorflow.Output{
            m.model.Graph.Operation("StatefulPartitionedCall").Output(0),
        },
        nil,
    )
    if err != nil {
        return nil, err
    }
	/*
    	DA CONTROLLARE COSA RITORNA
    */
    result := result[0].Value().([][]float32)

    // filter the actions
    for i, allowed := range actionFilter {
        if !allowed {
            result[i] = 0.0
        }
    }

    // return the index of highest value
    action := 0
    for i, value := range result {
        if value > maxValue {
            action = i
        }
    }
    return action, nil
}

func oneHotEncoding(list []string, str string) []int {
	indexMap := make(map[string]int)
	for i, v := range list {
		indexMap[v] = i
	}
	oneHot := make([]int, len(list))
	if idx, exists := indexMap[str]; exists {
		oneHot[idx] = 1
	}
	return oneHot
}

func getState_and_EdgeNode(r *scheduledRequest) ([]float64, string) {
	percAvailableLocalMemory = (node.Resources.AvailableMemMB + node.FreeableMemory(r.Fun))/node.Resources.MaxMemMB

	canExecuteOnEdge = true
	// url := pickBestEdgeNodeForOffloading(r)
	url := pickEdgeNodeForOffloading(r)
	if url == "" {
		canExecuteOnEdge = false
	}

	functionId = oneHotEncoding(r.Fun.GetAll(), r.Fun)

	classList := make([]string, 0, len(Classes))
    for key := range Classes {
        classList = append(classList, key)
    }
	classId = oneHotEncoding(classList, r.ClassService)

	state := []float64{percAvailableLocalMemory, canExecuteOnEdge}
	state = append(state, functionId...)
	state = append(state, classId...)
	state = append(state, !r.CanDoOffloading) // has_been_offloaded
	return state, url
}

func actionFilter(state []float64, r *scheduledRequest) []bool {
	actionFilter := []bool{true, true, true, true}
	availableMemory := node.Resources.MaxMemMB * state[0]
	canExecuteLocally := canExecute(r.Fun)
	canExecuteOnEdge := state[1] != 0 && state[len(state)-1] == 0	// canExecuteOnEdge && CanDoOffloading
	if !canExecuteLocally {
		actionFilter[0] = false
	}
	if !canExecuteOnEdge {
		actionFilter[2] = false
	}
	return actionFilter
}

func (d *decisionEngineDQN) Decide(r *scheduledRequest) (int, string, error) {

    /*
    STATE:
    - percAvailableLocalMemory (float)
    - canExecuteOnEdge (boolean)
    - functionId (one_hot)
    - classId (one_hot)
    - has_been_offloaded (boolean) == !CanDoOffloading (do not remove the ! cause the NN has been trained with has_been_offloaded)
    */

	name := r.Fun.Name
	class := r.ClassService

	arrivalChannel <- arrivalRequest{r, class.Name}

	/*
		CONTROLLARE SE E' CORRETTO FreeableMemory
	*/
	// retrieve state and the possible best edge node to offload to
	state, urlEdgeNode = getState_and_EdgeNode(r)

	// boolean list of allowed actions
	actionFilter = actionFilter(state, r)

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
    	action, err := model.Predict(state, actionFilter)
	    if err != nil {
	        log.Println("Error predicting:", err)
	        return nil, nil err
	    }
    }

    // map simulator action to Serverledge
    //  - simulator:   LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
    //  - Serverledge: DROP(0)-LOCAL(1)-CLOUD(2)-EDGE(3)
    action = (action + 1) % 4

    //log.Println("Action: %d", action)

	if action == DROP_REQUEST {
		requestChannel <- completedRequest{
			scheduledRequest: r,
			dropped:          true,
		}
	}
	return action, urlEdgeNode, nil
}

func (d *decisionEngineDQN) InitDecisionEngine() {
	// model initialization
    modelPath := "path/to/saved_model"
	// model, err := NewModel(modelPath)
    model, err := LoadModel(modelPath)
    if err != nil {
        log.Println("Error loading model: %v", err)
    }
	/*
    VEDERE SE QUESTO VA MESSE E NEL CASO VA QUA O A FINE RUN
    "viene chiamata solo al termine della funzione che contiene la dichiarazione defer"
    quindi in teoria così viene chiuso a fine InitDecisionEngine, cosa che non voglio
    */
   	// defer model.model.Session.Close()

	/*
    input := []float32{} // Inserisci i tuoi dati di input qui
    output, err := model.Predict(input)
    if err != nil {
        log.Println("Error predicting:", err)
        return
    }
    */
}

func (d *decisionEngineDQN) Completed(r *scheduledRequest, offloaded int) {
	// FIXME AUDIT log.Println("COMPLETED: in decisionEngineDQN")
	requestChannel <- completedRequest{
		scheduledRequest: r,
		location:         offloaded,
		dropped:          false,
	}
}

func (d *decisionEngineFlux) GetGrabber() metricGrabber {
	// VEDERE COSA DEVO FARCI
}
package scheduling

import (
	"log"
	"strings"
	"os"
	"encoding/json"
)

type decisionEngineProbabilistic struct {
	mg *metricGrabberDQN
}

var probabilities map[string]map[string][]float64

func (d *decisionEngineProbabilistic) Decide(r *scheduledRequest) int {
	function := r.Fun.Name
	class := r.ClassService.Name

	var pL float64
	var pC float64
	var pE float64
	var pD float64

	if classes, ok := probabilities[function]; ok {
        if probs, ok := classes[class]; ok {
            log.Printf("Probabilità per %s e classe %s: %v\n", function, class, probs)
			pL = probs[0]
			pC = probs[1]
			pE = probs[2]
			pD = probs[3]
        } else {
            log.Printf("Classe %s non trovata per funzione %s\n", class, function)
        }
    } else {
        log.Printf("Funzione %s non trovata\n", function)
    }

	if policyFlag == "edgeCloud" {
		// Cloud and Edge offloading allowed
		if !r.CanDoOffloading {
			// Can be executed only locally or dropped
			pD = pD / (pD + pL)
			pL = pL / (pD + pL)
			pC = 0
			pE = 0
		} else if !canExecute(r.Fun) {
			// Node can't execute function locally
			if pD == 0 && pC == 0 && pE == 0 {
				pD = 0
				pC = 0.5
				pE = 0.5
				pL = 0
			} else {
				pD = pD / (pD + pC + pE)
				pC = pC / (pD + pC + pE)
				pE = pE / (pD + pC + pE)
				pL = 0
			}
		}
	}

	// if !canAffordCloudOffloading(r) {
	// 	pC = 0
	// 	if pL == 0 && pE == 0 && pD == 0 {
	// 		pL = 0
	// 		pE = 0
	// 		pD = 1
	// 	} else {
	// 		pL = pL / (pL + pE + pD)
	// 		pE = pE / (pL + pE + pD)
	// 		pD = pD / (pL + pE + pD)
	// 	}
	// }

	//log.Printf("Probabilities after evaluation for %s-%s are pL:%f pE:%f pC:%f pD:%f", name, class.Name, pL, pE, pC, pD)

	prob := rGen.Float64()
	//log.Printf("prob: %f", prob)
	if prob <= pL {
		//log.Println("Execute LOCAL")
		return LOCAL_EXEC_REQUEST
	} else if prob <= pL+pE {
		//log.Println("Execute EDGE OFFLOAD")
		return EDGE_OFFLOAD_REQUEST
	} else if prob <= pL+pE+pC {
		//log.Println("Execute CLOUD OFFLOAD")
		return CLOUD_OFFLOAD_REQUEST
	} else {
		//log.Println("Execute DROP")
		d.mg.addStats(r,true,false)
		return DROP_REQUEST
	}
}

func (d *decisionEngineProbabilistic) InitDecisionEngine() {
	// Initializing probabilities
	probFilePath := "probs.json"

	file, err := os.Open(probFilePath)
    if err != nil {
        log.Println("Errore nell'apertura del file:", err)
        return
    }
    defer file.Close()

    // Mappa per decodificare il JSON
    var data map[string][]float64
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&data)
    if err != nil {
        log.Println("Errore nella decodifica JSON:", err)
        return
    }

    // Mappa annidata
    probabilities = make(map[string]map[string][]float64)

    // Riempie la mappa annidata
    for key, values := range data {
        // Separare la funzione e la classe dalla chiave usando strings.Split
        parts := strings.Split(key, "_")
        if len(parts) != 2 {
            log.Println("Chiave non valida:", key)
            continue
        }

        function := parts[0]
        class := parts[1]

        // Inizializzare la mappa per la funzione se non esiste
        if _, exists := probabilities[function]; !exists {
            probabilities[function] = make(map[string][]float64)
        }

        // Aggiungere i valori alla mappa della classe
        probabilities[function][class] = values
    }

    // Stampare il risultato
    log.Println("Mappa annidata:")
    for function, classes := range probabilities {
        log.Printf("%s: {\n", function)
        for class, values := range classes {
            log.Printf("  %s: %v\n", class, values)
        }
        log.Println("}")
    }
    
	d.mg = InitMG()
}

func (d *decisionEngineProbabilistic) Completed(r *scheduledRequest, offloaded int) {
	// FIXME AUDIT log.Println("COMPLETED: in decisionEngineProbabilistic")
	offloadDrop := offloaded != 0
	d.mg.addStats(r,false,offloadDrop)
}

func (d *decisionEngineProbabilistic) GetGrabber() metricGrabber {
	return nil
}

package kube

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"time"
        "fmt"
)

type EnhancedEvent struct {
	corev1.Event   `json:",inline"`
	InvolvedObject EnhancedObjectReference `json:"involvedObject"`
}

// Using an alias for map to allow overloading MarshalJSON. It is needed for some sinks to make
// output JSON compatible with the external system, e.g. BigQuery.
// TODO(vsbus): find a way to customize Map encoder externally.
type Map map[string]string
func (m Map) MarshalJSON() ([]byte, error) {
    type KV struct {
	  Key string
	  Value string
    }
    var s []KV
    for key, value := range m {
        s = append(s, KV{Key: key, Value: value})
    }

    return json.Marshal(s)
}

type EnhancedObjectReference struct {
	corev1.ObjectReference `json:",inline"`
	Labels      Map `json:"labels,omitempty"`
	Annotations Map `json:"annotations,omitempty"`
}


// TODO: 1. fix-bad-keys; 2. remove custom marshal above (or keep it?)... ; 3. find good name; 4. clean up code
func CopyMap(m map[string]interface{}) map[string]interface{} {
    cp := make(map[string]interface{})
    for k, v := range m {
        if v != nil {
            vm, ok := v.(map[string]interface{})
            if ok {
	        cp[k] = CopyMap(vm)
            } else {
                cp[k] = v
            }
        }
    }

    return cp
}



// ToJSON does not return an error because we are %99 confident it is JSON serializable.
// TODO(makin) Is it a bad practice? It's open to discussion.
func (e *EnhancedEvent) ToJSON() []byte {
	//Simple Employee JSON object which we will parse
	empJson := `{
		"id": 11,
		"name": "Irshad",
		"department": "IT",
		"designation": "Product Manager",
		"address": {
			"city": "Mumbai",
			"state": null,
			"country": "India"
		}
	}`

	// Declared an empty interface
	var result map[string]interface{}

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(empJson), &result)

	address := result["address"].(map[string]interface{})

	//Reading each value by its key
	fmt.Println("Id :", result["id"],
		"\nName :", result["name"],
		"\nDepartment :", result["department"],
		"\nDesignation :", result["designation"],
		"\nAddress :", address["city"], address["state"], address["country"])
		
	json_bytes, _ := json.Marshal(result)
	fmt.Println(string(json_bytes))

        // panic(nil)

	b, _ := json.Marshal(e)
        fmt.Println(string(b))
	return b
}

func (e *EnhancedEvent) GetTimestampMs() int64 {
	return e.FirstTimestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

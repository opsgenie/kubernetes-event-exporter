package sinks

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/clbanning/mxj"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

func GetString(event *kube.EnhancedEvent, text string) (string, error) {
	tmpl, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(text)
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	// TODO: Should we send event directly or more events?
	err = tmpl.Execute(buf, event)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func convertLayoutTemplate(layout map[string]interface{}, ev *kube.EnhancedEvent) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range layout {
		m, err := convertTemplate(value, ev)
		if err != nil {
			return nil, err
		}
		result[key] = m
	}
	return result, nil
}

func convertTemplate(value interface{}, ev *kube.EnhancedEvent) (interface{}, error) {
	switch v := value.(type) {
	case string:
		rendered, err := GetString(ev, v)
		if err != nil {
			return nil, err
		}

		return rendered, nil
	case map[interface{}]interface{}:
		strKeysMap := make(map[string]interface{})
		for k, v := range v {
			res, err := convertTemplate(v, ev)
			if err != nil {
				return nil, err
			}
			// TODO: It's a bit dangerous
			strKeysMap[k.(string)] = res
		}
		return strKeysMap, nil
	case map[string]interface{}:
		strKeysMap := make(map[string]interface{})
		for k, v := range v {
			res, err := convertTemplate(v, ev)
			if err != nil {
				return nil, err
			}
			strKeysMap[k] = res
		}
		return strKeysMap, nil
	case []interface{}:
		listConf := make([]interface{}, len(v))
		for i := range v {
			t, err := convertTemplate(v[i], ev)
			if err != nil {
				return nil, err
			}
			listConf[i] = t
		}
		return listConf, nil
	}
	return nil, nil
}

func serializeEventWithLayout(layout map[string]interface{}, ev *kube.EnhancedEvent) ([]byte, error) {
	var toSend []byte
	if layout != nil {
		res, err := convertLayoutTemplate(layout, ev)
		if err != nil {
			return nil, err
		}

		toSend, err = json.Marshal(res)
		if err != nil {
			return nil, err
		}
	} else {
		toSend = ev.ToJSON()
	}
	return toSend, nil
}

func serializeXMLEventWithLayout(layout map[string]interface{}, ev *kube.EnhancedEvent) ([]byte, error) {
	var (
		err     error
		toSend  []byte
		obj     interface{}
		rootTag string = mxj.DefaultRootTag
	)

	if layout != nil {
		res, err := convertLayoutTemplate(layout, ev)
		if err != nil {
			return nil, err
		}

		if len(res) == 1 {

			for k, v := range res {
				rootTag = k
				obj = v
			}
		} else {

			obj = res
		}

	} else {
		// convert event to map
		// TODO: create benchmark to check if reflect version is more performant
		mrv := make(map[string]interface{})
		jsonRepr, err := json.Marshal(ev)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonRepr, &mrv)
		if err != nil {
			return nil, err
		}
		obj = mrv
	}

	toSend, err = mxj.AnyXml(obj, rootTag)
	if err != nil {
		return nil, err
	}

	return []byte(xml.Header + string(toSend)), nil
}

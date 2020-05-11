package dynamic

type Ldap struct {
	Server  map[string]interface{}
	Entries []map[string]interface{}
}

// func (s *LdapServer) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	s.Config = make(map[string][]string)

// 	data := make(map[string]string)
// 	unmarshal(data)

// 	for k, v := range data {
// 		if strings.ToLower(k) == "listen" {
// 			s.Listen = v
// 		} else {
// 			list := make([]string, 1)
// 			list[0] = v
// 			s.Config[k] = list
// 		}
// 	}

// 	listData := make(map[string][]string)
// 	unmarshal(listData)

// 	for k, v := range listData {
// 		s.Config[k] = v
// 	}

// 	return nil
// }

// func (e *Entry) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	e.Attributes = make(map[string][]string)

// 	data := make(map[string]interface{})
// 	unmarshal(data)

// 	for k, v := range data {
// 		if s, ok := v.(string); ok {
// 			if strings.ToLower(k) == "dn" {
// 				e.Dn = s
// 				parts := strings.Split(s, ",")
// 				if len(parts) > 0 {
// 					kv := strings.Split(parts[0], "=")
// 					if len(kv) == 2 && strings.ToLower(kv[0]) == "cn" {
// 						e.Attributes["cn"] = []string{strings.TrimSpace(kv[1])}
// 					}
// 				} else {
// 					e.Attributes[k] = []string{s}
// 				}
// 			}
// 		} else if a, ok := v.([]interface{}); ok {
// 			list := make([]string, 0)
// 			for _, e := range a {
// 				if s, ok := e.(string); ok {
// 					list = append(list, s)
// 				}
// 			}
// 			e.Attributes[k] = list
// 		} else if ext, ok := v.(map[interface{}]interface{}); ok {

// 		} else {
// 			return fmt.Errorf("Unsupported configuration found near %v", k)
// 		}
// 	}

// 	func readExt(ext map[interface{}]interface{}){
// 		for p, o := range ext {
// 			switch strings.ToLower(p) {
// 			case "file":
// 				data, error := ioutil.ReadFile(o)
// 				if error != nil {
// 					log.WithFields(log.Fields{"Error": error, "Filename": file}).Error("error reading file")
// 					return
// 				}
// 			}
// 		}
// 	}

// data := make(map[string]string)
// unmarshal(data)

// for k, v := range data {
// 	if strings.ToLower(k) == "dn" {
// 		e.Dn = v
// 		parts := strings.Split(v, ",")
// 		if len(parts) > 0 {
// 			kv := strings.Split(parts[0], "=")
// 			if len(kv) == 2 && strings.ToLower(kv[0]) == "cn" {
// 				e.Attributes["cn"] = []string{strings.TrimSpace(kv[1])}
// 			}
// 		}
// 	} else {
// 		list := make([]string, 1)
// 		list[0] = v
// 		e.Attributes[k] = list
// 	}
// }

// listData := make(map[string][]string)
// unmarshal(listData)

// for k, v := range listData {
// 	e.Attributes[k] = v
// }

// t := ext["thumbnailphoto"]
// if t != nil {
// 	f := t.(map[string]interface{})["file"]

// 	fmt.Print(f)
// }

//return nil
//}

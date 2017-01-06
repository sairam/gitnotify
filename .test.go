package main

// type Setting struct {
// 	Version string                 `yaml:"version"`
// 	Data    map[string]Information `yaml:"fetched_info"`
// }
//
// type Information struct {
// 	RepoInformation `yaml:",inline"`
// 	OrgInformation  `yaml:",inline"`
// 	Type            string `yaml:"type"`
// }
//
// type OrgInformation struct {
// 	Name    string `yaml:"org_name,omitempty"`
// 	RefType string `yaml:"org_type,omitempty"`
// }
//
// type RepoInformation struct {
// 	Tags     []string    `yaml:"tags,omitempty"`
// 	Branches []string    `yaml:"branches,omitempty"`
// 	Commits  interface{} `yaml:"commits,omitempty"`
// }

// func (i *Information) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	var slice struct {
// 		Type string `yaml:"type"`
// 	}
// 	unmarshal(&slice)
// 	if slice.Type == "test" {
// 		var r RepoInformation
// 		unmarshal(&r)
// 		fmt.Println(r)
// 		var a Information
// 		a.RepoInformation = r
// 		a.Type = "repo"
// 		*i = a
// 		// i.Data = &r
// 	} else {
// 		r := OrgInformation{"asiram", "user"}
// 		var a Information
// 		a.OrgInformation = r
// 		a.Type = "org"
// 		*i = a
// 		// unmarshal(&r)
//
// 	}
// 	return nil
// }
//

// func main() {
// 	c := &Setting{}
// 	data, err := ioutil.ReadFile("data/github/sairam/settings.yml")
// 	if os.IsNotExist(err) {
// 		fmt.Println(err)
// 	}
//
// 	err = yaml.Unmarshal(data, c)
// 	fmt.Println(err)
// 	// t := c.Data["minio/minio"].(RepoInformation)
// 	// fmt.Println(t)
// 	// t, _ := yaml.Marshal(c.Data["minio/minio"])
// 	// r := &RepoInformation{}
// 	// yaml.Unmarshal(t, r)
// 	// fmt.Println(r.Branches)
//
// 	fmt.Println(reflect.TypeOf(c.Data["avelino/awesome-go"].OrgInformation))
// 	fmt.Println(c.Data["avelino/awesome-go"])
//
// 	out, err := yaml.Marshal(c)
// 	fmt.Println(err)
// 	fmt.Println(string(out))
//
// }

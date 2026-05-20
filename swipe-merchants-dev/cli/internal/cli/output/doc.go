// Package output renders command results as table, json, or yaml. Tables
// use go-pretty; JSON uses encoding/json; YAML uses gopkg.in/yaml.v3.
// Every command in the tree gets a Renderer via the global --output flag.
package output

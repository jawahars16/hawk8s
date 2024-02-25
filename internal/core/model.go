package core

import (
	"hash/fnv"
)

type (
	nodeViewModel struct {
		Nodes           []node
		ActiveNamespace string
		ActiveMode      string
		Title           string
		Namespaces      []namespace
		Modes           []mode
		Error           string
	}
	podViewModel struct {
		Pods  []pod
		Error string
	}
	node struct {
		Name string
		Info string
		Pods []pod
	}
	pod struct {
		Name        string
		CpuSize     string
		MemorySize  string
		Color       string
		Namespace   string
		Status      string
		CpuUsage    string
		MemoryUsage string
	}
	mode struct {
		Name  string
		Value string
	}
	namespace struct {
		Name  string
		Color string
	}
)

func namespaceByName(name string) namespace {
	return namespace{
		Name:  name,
		Color: namespaceColors[hash(name)%uint32(len(namespaceColors))],
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

var namespaceColors = []string{
	"#E91E63", // Pink 500
	"#9C27B0", // Purple 500
	"#673AB7", // Deep Purple 500
	"#3F51B5", // Indigo 500
	"#2196F3", // Blue 500
	"#03A9F4", // Light Blue 500
	"#00BCD4", // Cyan 500
	"#009688", // Teal 500
	"#4CAF50", // Green 500
	"#8BC34A", // Light Green 500
	"#CDDC39", // Lime 500
	"#FFEB3B", // Yellow 500
	"#FFC107", // Amber 500
	"#FF9800", // Orange 500
	"#FF5722", // Deep Orange 500
	"#795548", // Brown 500
	"#9E9E9E", // Grey 500
	"#607D8B", // Blue Grey 500
	"#EC407A", // Pink 300
	"#AB47BC", // Purple 300
	"#7E57C2", // Deep Purple 300
	"#5C6BC0", // Indigo 300
	"#42A5F5", // Blue 300
	"#29B6F6", // Light Blue 300
	"#26C6DA", // Cyan 300
	"#26A69A", // Teal 300
	"#66BB6A", // Green 300
	"#9CCC65", // Light Green 300
	"#D4E157", // Lime 300
	"#FFEE58", // Yellow 300
	"#FFCA28", // Amber 300
	"#FFA726", // Orange 300
	"#FF7043", // Deep Orange 300
	"#8D6E63", // Brown 300
	"#BDBDBD", // Grey 300
	"#78909C", // Blue Grey 300
}

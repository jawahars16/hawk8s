package kubeclient

type (
	Node struct {
		Name              string
		Status            string
		AllocatableMemory int64
		TotalMemory       int64
		AvailableCPU      int64
	}

	Pod struct {
		Name        string
		Node        string
		Namespace   string
		MemoryUsage int64
		CPUUsage    int64
		Status 	string
	}
)

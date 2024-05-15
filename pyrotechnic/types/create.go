package main

type CreateRequest struct {
	RootDrivePath string `json:"root_image_path"`
	KernelPath    string `json:"kernel_path"`
	CloudInitPath string `json:"cloud_init_path"`
}

type CreateResponse struct {
	IpAddress string `json:"ip_address"`
	ID        string `json:"id"`
}


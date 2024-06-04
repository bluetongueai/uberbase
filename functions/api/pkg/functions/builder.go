package functions

import (
	"context"
	"log"
	"os"
	"strings"
)

type builderConfig struct {
	ImageDirPath string
}

func buildImages(client client, config builderConfig) error {
	imagesToBuild, err := getAllImagesFromDir(config.ImageDirPath)
	if err != nil {
		panic(err)
		return err
	}
	log.Printf("building images: %v", imagesToBuild)
	for _, image := range imagesToBuild {
		imageName := imageNameFromDockerfile(image)
		buildImage(client, config.ImageDirPath, "bluetongueai/functions-"+imageName+":latest")
	}
	log.Println("images built")
	return nil
}

func getAllImagesFromDir(path string) ([]string, error) {
	// get all subdirs of path
	subdirs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	// get all names of subdirs
	var images []string
	for _, subdir := range subdirs {
		images = append(images, subdir.Name())
	}
	return images, nil
}

func imageNameFromDockerfile(dockerfile string) string {
	return strings.Split(dockerfile, "/")[0]
}

func buildImage(client client, basePath string, imageName string) error {
	ctx := context.Background()
	image := basePath + "/" + imageName + "/Dockerfile"
	buildContext := basePath + "/" + imageName + "/"
	err := client.Build(ctx, imageName+":latest", image, buildContext)
	if err != nil {
		return err
	}
	return nil
}

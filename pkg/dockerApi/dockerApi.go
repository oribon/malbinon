package dockerapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func PullImage(imageName string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()

	authConfig := types.AuthConfig{
		Username: "",
		Password: "",
	}
	encodedJSON, err := json.Marshal(authConfig)
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	pullOptions := types.ImagePullOptions{RegistryAuth: authStr}
	rc, err := cli.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	imageWriter := bytes.NewBufferString("")
	_, err = io.Copy(imageWriter, rc)
	if err != nil {
		return "", err
	}
	r, _ := regexp.Compile("for (.*)\"")
	regexMatches := r.FindStringSubmatch(imageWriter.String())
	pulledImageName := regexMatches[len(regexMatches)-1]

	return pulledImageName, nil
}

func SaveImage(imageName string, baseDir string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	replacer := strings.NewReplacer("/", "_", ":", "..")
	formattedImageName := replacer.Replace(imageName)

	imageFile, err := os.Create(filepath.Join(baseDir, formattedImageName))
	if err != nil {
		return "", err
	}

	rc, err := cli.ImageSave(ctx, []string{imageName})
	if err != nil {
		return "", err
	}

	_, err = io.Copy(imageFile, rc)
	if err != nil {
		return "", err
	}
	return formattedImageName, nil
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
	}
	defer cli.Close()

	authConfig := types.AuthConfig{
		Username: "",
		Password: "",
	}
	encodedJSON, err := json.Marshal(authConfig)
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	pullOptions := types.ImagePullOptions{RegistryAuth: authStr}
	rc, err := cli.ImagePull(ctx, "library/busybox", pullOptions)
	if err != nil {
		fmt.Println(err)
	}
	defer rc.Close()

	imageWriter := bytes.NewBufferString("")
	_, err = io.Copy(imageWriter, rc)
	if err != nil {
		fmt.Println(err)
	}
	r, _ := regexp.Compile("for (.*)\"")
	regexMatches := r.FindStringSubmatch(imageWriter.String())
	imageName := regexMatches[len(regexMatches)-1]
	replacer := strings.NewReplacer("/", "_", ":", "..")
	cleanImageName := replacer.Replace(imageName)
	fmt.Println(imageName, cleanImageName)

	imageFile, err := os.Create("/tmp/" + cleanImageName)
	if err != nil {
		fmt.Println(err)
	}

	rc, err = cli.ImageSave(ctx, []string{imageName})
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(imageFile, rc)
	if err != nil {
		fmt.Println(err)
	}
}

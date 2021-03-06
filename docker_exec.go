package main

import (
	"errors"
	"log"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DockerCP Copy local files to the docker container，or docker to local
// usage: docker cp test.go [container ID or Name]:/
// example: docker cp test.go gocode11:/
func DockerCP(user *User) error {
	var c string
	switch user.Code.Type {
	case TypeGO, TypeJava, TypeC, TypeCPP:
		// docker cp /Users/abc/GolandProjects/tools/codeonline/code/1234.go gotest:code/
		c = "docker cp " + TempFilePath + user.Filename + " " + user.ContainerName + ":" + SourceFilePath
	default:
		return errors.New("unknown code type")
	}
	_, err := dockerExec(user, c)
	if err != nil {
		return err
	}
	return nil
}

// DockerRun create a container
func DockerRun(user *User) error {
	var c string
	rand.Seed(time.Now().UnixNano())
	containerName := strconv.Itoa(rand.Intn(10000))
	user.ContainerName = containerName
	switch user.Code.Type {
	case TypeGO:
		// docker run -d -it --name gocode11 golang
		c = "docker run -d -it --name " + containerName + " " + ImageGo
	case TypeJava:
		c = "docker run -d -it --name " + containerName + " " + ImageJava
	case TypeC:
		c = "docker run -d -it --name " + containerName + " " + ImageC
	case TypeCPP:
		c = "docker run -d -it --name " + containerName + " " + ImageCPP
	default:
		return errors.New("unknown code type")
	}
	_, err := dockerExec(user, c)
	if err != nil {
		return err
	}
	return nil
}

// DockerExecAndRunCode exec container and execution this user's code, and return the result
// Example: docker exec -it gocode11 sh -c " ls -l && go run test.go"
func DockerExecAndRunCode(user *User) ([]byte, error) {
	var c string
	switch user.Code.Type {
	case TypeGO:
		// docker exec -i gotest sh -c "go run 1420.go"
		c = "docker exec -i " + user.ContainerName +
			` sh -c "cd ../code && ` + "go run " + user.Filename + `"`
	case TypeJava:
		// java 需要先 javac xxx.java 编译出 class 文件，再使用 java xxx（不要后缀名）获得运行结果
		index := strings.Index(user.Filename, ".")
		// 去除 .class 后缀名
		noSuffix := user.Filename[:index]
		c = "docker exec -i " + user.ContainerName +
			` sh -c "cd ../code && ` + "javac " + user.Filename +
			" && " + "java " + noSuffix + `"`
	case TypeC:
		// gcc 编译出二进制文件，再执行，二进制文件是没有后缀名的，在这里去除
		index := strings.Index(user.Filename, ".")
		noSuffix := user.Filename[:index]
		// gcc -o test test.c && ./test
		c = "docker exec -i " + user.ContainerName +
			` sh -c "cd ../code && ` + "gcc -o " + noSuffix + " " +
			user.Filename + " && " + "./" + noSuffix + `"`
	case TypeCPP:
		index := strings.Index(user.Filename, ".")
		noSuffix := user.Filename[:index]
		// gcc 更改为 g++
		c = "docker exec -i " + user.ContainerName +
			` sh -c "cd ../code && ` + "g++ -o " + noSuffix + " " +
			user.Filename + " && " + "./" + noSuffix + `"`
	default:
		return nil, errors.New("unknown code type")
	}
	res, err := dockerExec(user, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// DockerExecAndRemoveFile exec container, and then remove this user's code file
func DockerExecAndRemoveFile(user *User) error {
	var c string
	switch user.Code.Type {
	case TypeGO:
		// docker exec and rm command:  docker exec -i gotest sh -c "cd ../code && rm 214.go"
		c = "docker exec -i " + user.ContainerName + ` sh -c "cd ../code && ` + "rm " + user.Filename + `"`
	default:
		return errors.New("unknown code type")
	}
	_, err := dockerExec(user, c)
	if err != nil {
		return err
	}
	return nil
}

// DockerExecAndCreateDir exec container and create dir "code" in root path "/"
// the user code is store in here
func DockerExecAndCreateDir(user *User) error {
	var c string
	switch user.Code.Type {
	case TypeGO, TypeJava, TypeC, TypeCPP:
		c = "docker exec -i " + user.ContainerName + ` sh -c "cd .. && mkdir code"`
	default:
		return errors.New("unknown code type")
	}
	_, err := dockerExec(user, c)
	if err != nil {
		return err
	}
	return nil
}

// DockerRM	remove the container according to the code type
func DockerRM(user *User) error {
	var c string
	switch user.Code.Type {
	case TypeGO, TypeJava, TypeC, TypeCPP:
		c = "docker rm -f " + user.ContainerName
	default:
		return errors.New("unknown code type")
	}
	_, err := dockerExec(user, c)
	if err != nil {
		return err
	}
	return nil
}

func dockerExec(user *User, cmd string) ([]byte, error) {
	if user.ContainerName == "" {
		log.Println("user container name is nil, you need create a container first")
		return nil, errors.New("user container name is nil, you need create a container first")
	}
	c := exec.Command("bash", "-c", cmd)
	log.Println("docker exec command: ", cmd)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Printf("docker exec [ " + cmd + " ] error: %v\n", err)
		return nil, errors.New(string(output))
	}
	log.Println("docker exec result: ", string(output))
	return output, nil
}

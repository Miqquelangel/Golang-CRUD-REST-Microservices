package main

import (
        "bufio"
        "fmt"
        "os"
        "os/exec"
        "strings"
)

func main() {
        reader := bufio.NewReader(os.Stdin)

        fmt.Print("Enter the port to expose the app: ")
        port, _ := reader.ReadString('\n')

        fmt.Print("Enter the name of the app: ")
        nameApp, _ := reader.ReadString('\n')

        dockerfile := 

`FROM golang:1.20
WORKDIR /app
COPY . .
RUN go build -o main .
EXPOSE ` + port + `
CMD ["./main"]`

        file, err := os.Create("Dockerfile")
        if err != nil {
                fmt.Println(err)
                return
        }
        defer file.Close()

        _, err = file.WriteString(dockerfile)
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Dockerfile created successfully")

        // Execute go mod init and go mod tidy
        cmd := exec.Command("go", "mod", "init", "test")
        err = cmd.Run()
        if err != nil {
                fmt.Println("k")
                return
        }

        cmd = exec.Command("go", "mod", "tidy")
        err = cmd.Run()
        if err != nil {
                fmt.Println("l")
                return
        }

        // Build the Docker image
        cmd = exec.Command("docker", "build", "-t", strings.TrimSpace(nameApp), ".")
        err = cmd.Run()
        if err != nil {
                fmt.Println(nameApp)
                return
        }

        fmt.Println("Docker image built successfully")

        // Ask the user for the repository and version
        fmt.Print("Enter the user/repository: ")
        repository, _ := reader.ReadString('\n')

        fmt.Print("Enter the version: ")
        version, _ := reader.ReadString('\n')

        // Execute docker tag
        cmd = exec.Command("docker", "tag", strings.TrimSpace(nameApp)+":latest", strings.TrimSpace(repository)+":"+strings.TrimSpace(version))
        err = cmd.Run()
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Docker image tagged successfully")

        // Execute docker push
        cmd = exec.Command("docker", "push", strings.TrimSpace(repository)+":"+strings.TrimSpace(version))
        err = cmd.Run()
        if err != nil {
                fmt.Println(strings.TrimSpace(repository))
                return
        }

        fmt.Sprintf("Docker image pushed into the repository %d completed successfully", repository)

        // Create the deployment YAML file
        deploymentYAML := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: %s:%s`, strings.TrimSpace(repository), strings.TrimSpace(version))

        err = os.WriteFile("deployment-cicd.yaml", []byte(deploymentYAML), 0644)
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Deployment YAML file created successfully")

        // Execute the deployment
        cmd = exec.Command("kubectl", "apply", "-f", "deployment-cicd.yaml")
        err = cmd.Run()
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Deployment executed successfully")

        // Create the service YAML file
        serviceYAML := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: app-service
spec:
  type: NodePort
  selector:
    app: app
  ports:
    - protocol: TCP
      port: 80
      targetPort: %s
      nodePort: 32222`, port)

        err = os.WriteFile("service-cicd.yaml", []byte(serviceYAML), 0644)
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Service YAML file created successfully")

        // Execute the deployment
        cmd = exec.Command("kubectl", "apply", "-f", "service-cicd.yaml")
        err = cmd.Run()
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Println("Service executed successfully")

}

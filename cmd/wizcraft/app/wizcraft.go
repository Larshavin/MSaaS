package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func NewWizcraftCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "wizcraft",
		Long: `

		`,
		Short: "Wizcraft is a CLI tool for managing microservices",
	}

	cmd.AddCommand(CreateProject())

	return cmd
}

func CreateProject() *cobra.Command {
	var appType string

	cmd := &cobra.Command{
		Use:   "create-project",
		Short: "Create a new project",
		Long: `Create a new project with the specified application type.
		Supported application types:
  		- spring: Create a Spring Boot project`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check if Spring CLI is installed
			if _, err := exec.LookPath("spring"); err != nil {
				fmt.Println("Spring CLI is not installed. Please follow the installation guide:")
				fmt.Println("https://docs.spring.io/spring-boot/installing.html")
				return
			}

			var name, javaVersion, projectDir string

			fmt.Print("Enter project name: ")
			fmt.Scanln(&name)

			if appType == "spring" {
				fmt.Print("Enter Java version (e.g., 17, 21): ")
				fmt.Scanln(&javaVersion)
			}

			fmt.Print("Enter project directory (default is current directory): ")
			fmt.Scanln(&projectDir)

			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}

			projectPath := filepath.Join(projectDir, name)

			var initCmd *exec.Cmd
			if appType == "spring" {
				initCmd = exec.Command("spring", "init", "--build=gradle", "--java-version="+javaVersion, "--name="+name, "--type=gradle-project-kotlin", projectPath)
			} else {
				fmt.Println("Unsupported application type:", appType)
				return
			}

			initCmd.Stdout = os.Stdout
			initCmd.Stderr = os.Stderr

			if err := initCmd.Run(); err != nil {
				fmt.Println("Error initializing project:", err)
				return
			}

			fmt.Println("Project initialized successfully in directory:", projectPath)

			if appType == "spring" {
				// Run Gradle build using ./gradlew
				buildCmd := exec.Command("./gradlew", "build")
				buildCmd.Dir = projectPath
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr

				if err := buildCmd.Run(); err != nil {
					fmt.Println("Error building project:", err)
					return
				}

				fmt.Println("Project built successfully")

				// Create Dockerfile
				dockerfilePath := filepath.Join(projectPath, "Dockerfile")
				dockerfileContent := fmt.Sprintf(`FROM openjdk:%s-jdk-slim
					WORKDIR /app
					COPY build/libs/*.jar app.jar
			ENTRYPOINT ["java", "-jar", "app.jar"]
			`, javaVersion)
				if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
					fmt.Println("Error creating Dockerfile:", err)
					return
				}

				fmt.Println("Dockerfile created successfully in directory:", projectPath)
			}

			// Build Docker image
			imageName := fmt.Sprintf("%s:latest", name)
			dockerBuildCmd := exec.Command("docker", "build", "-t", imageName, projectPath)
			dockerBuildCmd.Stdout = os.Stdout
			dockerBuildCmd.Stderr = os.Stderr

			if err := dockerBuildCmd.Run(); err != nil {
				fmt.Println("Error building Docker image:", err)
			} else {
				fmt.Println("Docker image built successfully with name:", imageName)
			}
		},
	}

	cmd.Flags().StringVar(&appType, "app", "", "Specify the application type (spring/next)")
	cmd.MarkFlagRequired("app")

	return cmd
}

package performance //nolint:testpackage

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	pushedImagesNames []string
	start             time.Time
	op                skopeoOp
	configPath        string
	serverConfig      *ServerConfig
)

func init() {
	flag.StringVar(&configPath, "server.config", "",
		"path to the server config file")
}

func TestPerformance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Suite")
}

// numarul de repetari din fisierul de config

var _ = BeforeSuite(func() {
	err := os.MkdirAll("pulled_images", 0777)
	if err != nil {
		log.Fatalln("Error creating pulled_images dir!", err)
	}

	Expect(configPath).To(BeAnExistingFile(),
		"Invalid test suite argument. server.config should be an existing file.")

	serverConfig = LoadConfig(configPath)
	op = skopeoOp{
		serverConfig.Username,
		serverConfig.Password,
		serverConfig.Address,
		serverConfig.TlsVerify,
		serverConfig.Repo,
	}
})

var _ = AfterSuite(func() {
	err := os.RemoveAll("pulled_images")
	if err != nil {
		log.Fatalln("Error removing pulled_images dir!", err)
	}
})

var _ = Describe("Check Zot Performance", func() {
	It("skopeo should be installed", func() {
		Expect(checkSkopeoBinary()).To(Equal(true))
	})

	// Measure("it should do something hard efficiently", func(b Benchmarker) {
	// runtime := b.Time("runtime", func() {
	It("Push operation", func() {
		imageName := "zot-tests-dummy-push"
		Expect(runPushCommand(op, imageName)).To(Equal(true))
	})

	It("Pull operation", func() {
		imageName := "zot-tests-dummy-push"
		Expect(runCopy(op, imageName, imageName, false)).To(Equal(true))
	})

	It("Delete operation", func() {
		imageName := "zot-tests-dummy-push"
		Expect(runDeleteCommand(op, imageName)).To(Equal(true))
	})
	// 	})

	// 	Expect(runtime.Seconds()).To(BeNumerically("<", 2),
	// 		"Push and Pull oprations shouldn't take too long!")
	// }, 10)

	// Measure("it should do something hard efficiently", func(b Benchmarker) {
	// 	runtime := b.Time("runtime", func() {
	It("Running multiple images parallel write commands", func() {
		imageNameIdx := "zot-tests-parallel-images-dummy-%d"
		var commands []string
		var pushedImages []string

		for i := 1; i <= 5; i++ {
			imageName := fmt.Sprintf(imageNameIdx, i)
			pushedImages = append(pushedImages, imageName)
			arguments := setCopyCommand(op)

			arguments = setCommandVariables(op, arguments, imageName, imageName, true)
			commands = append(commands, strings.Join(arguments, " "))
		}

		Expect(runCommands(commands)).To(Equal(true))
		pushedImagesNames = append(pushedImagesNames, pushedImages...)
	})

	It("Running multiple images parallel read commands", func() {
		imageNameIdx := "zot-tests-parallel-images-dummy-%d"
		var commands []string

		for i := 1; i <= 5; i++ {
			imageName := fmt.Sprintf(imageNameIdx, i)
			arguments := setCopyCommand(op)
			arguments = setCommandVariables(op, arguments, imageName, imageName, false)
			commands = append(commands, strings.Join(arguments, " "))
		}
		Expect(runCommands(commands)).To(Equal(true))
	})

	It("Running skopeo delete command", func() {
		var commands []string

		for _, imageName := range pushedImagesNames {
			arguments := setDeleteCommand(op, imageName)
			commands = append(commands, strings.Join(arguments, " "))
		}

		Expect(runCommands(commands)).To(Equal(true))
	})
	// 	})

	// 	Expect(runtime.Seconds()).To(BeNumerically("<", 2),
	// 		"Push and Pull oprations shouldn't take too long!")
	// }, 10)

	// Measure("it should do something hard efficiently", func(b Benchmarker) {
	// 	runtime := b.Time("runtime", func() {
	It("Running single image parallel write commands", func() {
		imageName := "zot-tests-single-images-dummy"
		var commands []string

		for i := 1; i <= 5; i++ {
			arguments := setCopyCommand(op)
			arguments = setCommandVariables(op, arguments, imageName, imageName, true)
			commands = append(commands, strings.Join(arguments, " "))
		}

		Expect(runCommands(commands)).To(Equal(true))

		pushedImagesNames = append(pushedImagesNames, imageName)
	})

	It("Running single image parallel read commands", func() {
		imageName := "zot-tests-single-images-dummy"
		var commands []string

		for i := 1; i <= 5; i++ {
			arguments := setCopyCommand(op)
			arguments = setCommandVariables(op, arguments, imageName,
				fmt.Sprintf("%s-%d:%s", imageName, i, "0.1.1"), false)
			commands = append(commands, strings.Join(arguments, " "))
		}
		Expect(runCommands(commands)).To(Equal(true))
	})
	// })

	It("Running skopeo delete command", func() {
		var commands []string

		for _, imageName := range pushedImagesNames {
			arguments := setDeleteCommand(op, imageName)
			commands = append(commands, strings.Join(arguments, " "))
		}

		Expect(runCommands(commands)).To(Equal(true))
	})

	// Expect(runtime.Seconds()).To(BeNumerically("<", 2),
	// 	"Push and Pull oprations shouldn't take too long!")
	// }, 1)

	It("Running skopeo delete command", func() {
		var commands []string

		for _, imageName := range pushedImagesNames {
			arguments := setDeleteCommand(op, imageName)
			commands = append(commands, strings.Join(arguments, " "))
		}

		Expect(runCommands(commands)).To(Equal(true))
	})
})

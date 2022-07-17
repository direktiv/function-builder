package main

import (
	"github.com/spf13/cobra"
)

// https://github.com/otiai10/copy

var prepCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the basic directory layout of a Direktiv function",
	RunE: func(cmd *cobra.Command, args []string) error {
		return prepare()
	},
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates the source code for the function",
	RunE: func(cmd *cobra.Command, args []string) error {
		return generate()
	},
}

var rootCmd = &cobra.Command{
	Use:   "service-builder",
	Short: "A source code generator for Direktiv functions",
	Long: `Direktiv's service builder can build functiopns based on the swagger specification.
It provides start templates and generates source code and a Docker image for the function.`,
}

func init() {

	prepCmd.Flags().StringVarP(&fnName, "function", "f", "", "Name of the function")
	prepCmd.MarkFlagRequired("function")
	prepCmd.Flags().StringVarP(&fnDir, "directory", "d", "",
		"Target directory. If not set a new directory with the name of the service will be created.")

	genCmd.Flags().StringVarP(&fnDir, "directory", "d", "",
		"Directory with the initialised Direktiv function")
	genCmd.MarkFlagRequired("directory")

	rootCmd.AddCommand(prepCmd)
	rootCmd.AddCommand(genCmd)

}

func main() {

	rootCmd.Execute()
	// var functionName string
	// flag.StringVar(&functionName, "function", "", "name of the function")
	// flag.Parse()

	// if functionName == "" {
	// 	log.Fatalln("function name is required, e.g. --function=myfunc")
	// }

	// fmt.Printf("HELLO %s<<<<", functionName)
}

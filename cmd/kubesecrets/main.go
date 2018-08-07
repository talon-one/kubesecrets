package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"encoding/base64"

	"github.com/talon-one/kubesecrets"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeConfigFlag    *string
	inClusterAuth     = kingpin.Flag("incluster", "use in-cluster authentication").Default("false").Bool()
	namespaceFlag     = kingpin.Flag("namespace", "namespace to use").Default("default").String()
	outputFlag        = kingpin.Flag("output", "output format to use").Default("json").Enum("json", "yaml")
	getCommand        = kingpin.Command("get", "get a secret").Alias("list")
	getCommandFilters = getCommand.Arg("filter", "").Strings()
	setCommand        = kingpin.Command("set", "set a secret")
	setBase64Flag     = setCommand.Flag("base64", "value is specified in base64 format").Default("false").Bool()
	setCommandName    = setCommand.Arg("name", "secret name").Required().String()
	setCommandValue   = setCommand.Arg("value", "secret value").Required().String()
	deleteCommand     = kingpin.Command("delete", "delete a secret")
	deleteCommandName = deleteCommand.Arg("name", "secret name").Required().String()
)

func main() {
	kubeConfig := kingpin.Flag("kubeconfig", "absolute path to the kubeconfig file")
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = kubeConfig.Default(filepath.Join(home, ".kube", "config"))
	}
	kubeConfigFlag = kubeConfig.String()
	kingpin.Version("1.0.0")
	os.Exit(Main())
}

func Main() int {
	cmd := kingpin.MustParse(kingpin.CommandLine.Parse(os.Args[1:]))

	var config *rest.Config
	if *inClusterAuth {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	} else {
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfigFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	switch cmd {
	case getCommand.FullCommand():
		if err := listSecrets(clientSet, (*getCommandFilters)...); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	case setCommand.FullCommand():
		if err := setSecret(clientSet, *setCommandName, *setCommandValue); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	case deleteCommand.FullCommand():
		if err := deleteSecret(clientSet, *deleteCommandName); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}
	return 0
}

func listSecrets(clientSet *kubernetes.Clientset, filter ...string) error {
	secrets, err := kubesecrets.GetSecrets(*namespaceFlag, clientSet, filter...)
	if err != nil {
		return err
	}
	return printSecrets(secrets)
}

func setSecret(clientSet *kubernetes.Clientset, name string, value string) error {
	var data []byte
	if *setBase64Flag {
		var err error
		data, err = base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
	} else {
		data = []byte(value)
	}
	secret, err := kubesecrets.SetSecret(*namespaceFlag, clientSet, name, data)
	if err != nil {
		return err
	}
	return printSecrets([]kubesecrets.Secret{secret})
}

func deleteSecret(clientSet *kubernetes.Clientset, name string) error {
	secret, err := kubesecrets.DeleteSecret(*namespaceFlag, clientSet, name)
	if err != nil {
		return err
	}
	return printSecrets([]kubesecrets.Secret{secret})
}

func printSecrets(secrets []kubesecrets.Secret) error {
	switch strings.ToLower(*outputFlag) {
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		if err := enc.Encode(secrets); err != nil {
			return err
		}
	default:
		if len(secrets) == 0 {
			fmt.Fprintln(os.Stdout, "[]")
			return nil
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		if err := enc.Encode(secrets); err != nil {
			return err
		}
	}
	return nil
}

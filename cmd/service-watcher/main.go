package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var version string = "unknown"

func main() {
	dumpSvc := flag.Bool("services", false, "Dump all services in json format")
	watchSvc := flag.Bool("watch", false, "Watch all services")
	script := flag.String("script", "", "Script called on a sevice update")
	ver := flag.Bool("version", false, "Print version and quit")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	if *dumpSvc {
		if err := dumpServices(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	if *watchSvc {
		if err := watchServices(*script); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	log.Println("Use -help")
	os.Exit(0)
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig :=
			clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func dumpServices() error {
	clientset, err := getClientset()
	if err != nil {
		return err
	}

	api := clientset.Core()
	svcs, err := api.Services("").List(meta.ListOptions{})
	if err != nil {
		return err
	}

	// Types in; k8s.io/kubernetes/pkg/apis/core/types.go
	for _, n := range svcs.Items {
		if s, err := json.Marshal(n); err == nil {
			fmt.Println(string(s))
		}
	}

	return nil
}

func watchServices(script string) error {
	clientset, err := getClientset()
	if err != nil {
		return err
	}
	api := clientset.Core()

	watcher, err := api.Services("").Watch(meta.ListOptions{})
	if err != nil {
		return err
	}

	var c []string
	if script != "" {
		c = strings.Split(script, " ")
	}

	ch := watcher.ResultChan()
	for {

		// Wait for any update
		<-ch

		// There will be an item in the channel for *every* service that
		// is updated. Especially on start-up there will be a storm of
		// events but we really want to update just once. So wait a short
		// time for more events and then drain the channel.
		time.Sleep(100 * time.Millisecond)
		for drained := false; drained == false; {
			select {
			case <-ch:
			default:
				drained = true
			}
		}

		if script == "" {
			log.Println("Some service altered...")
		} else {
			log.Println("Calling;", c)
			ctx, cancel := context.WithTimeout(
				context.Background(), 10*time.Second)
			defer cancel()
			err := exec.CommandContext(ctx, c[0], c[1:]...).Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

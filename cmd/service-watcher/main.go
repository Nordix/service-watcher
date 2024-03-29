package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var version string = "unknown"

func main() {
	dumpSvc := flag.Bool("services", false, "Dump all services in json format")
	watchSvc := flag.Bool("watch", false, "Watch all services")
	watchNodesFlag := flag.Bool("watch-nodes", false, "Watch nodes")
	script := flag.String("script", "", "Script called on api update")
	label := flag.String("label", "", "service-proxy-name")
	ver := flag.Bool("version", false, "Print version and quit")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	if *dumpSvc {
		if err := dumpServices(*label); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	if *watchSvc {
		if err := watchServices(*script, *label); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	if *watchNodesFlag {
		if err := watchNodes(*script); err != nil {
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

func dumpServices(label string) error {
	clientset, err := getClientset()
	if err != nil {
		return err
	}

	api := clientset.CoreV1()
	options := meta.ListOptions{}
	if label != "" {
		options.LabelSelector = "service.kubernetes.io/service-proxy-name=" + label
	}
	svcs, err := api.Services("").List(context.TODO(), options)
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

func watchServices(script string, label string) error {
	clientset, err := getClientset()
	if err != nil {
		return err
	}
	api := clientset.CoreV1()
	options := meta.ListOptions{}
	if label != "" {
		options.LabelSelector = "service.kubernetes.io/service-proxy-name=" + label
	}
	watcher, err := api.Services("").Watch(context.TODO(), options)
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

func watchNodes(script string) error {
	clientset, err := getClientset()
	if err != nil {
		return err
	}
	api := clientset.CoreV1()

	watcher, err := api.Nodes().Watch(context.TODO(), meta.ListOptions{})
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
		event := <-ch
		if event.Type == watch.Error {
			log.Fatal("Watch error")
		}

		// Reduce burstiness
		time.Sleep(1000 * time.Millisecond)
		for drained := false; drained == false; {
			select {
			case <-ch:
			default:
				drained = true
			}
		}

		if script == "" {
			log.Printf("Some nodes altered [%s]\n", event.Type)
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

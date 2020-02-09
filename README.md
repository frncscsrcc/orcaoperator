ORCA: In cluster orchestrator.
===

Motivations
---
This project is the result of (several) coding hours. The real reasons why I decided to dedicate all this time to a side project like this, were the fact that I would like to improve my skills in Go and the fact I would like to understand the internals of kubernetes, being able to understand the source code. I considered the development of kube-operators the best way to dig inside these two big topics.

Disclaimer
---
The current state of the project is at PoC level. This does not mean it is a buggy product. Actually all the functionality described are properly working. What is missing is a complete automatic test suite. At the moment I don't consider this product ready for production, so use it at your own risk!


Introduction
---
**Orca** is a kubernetes operator designed to run inside your cluster. The purpose of this application is to run Tasks based on simple kubernetes configurations, that define work-flows.

In order to use Orca, you need to define two new Kubernetes resources in your cluster: **tasks** and **ignitors**.

You can imagine a task as a simple Pod that, based on certain conditions, is created and executed to completion (a kind of kubernetes Job). An ignitor is just an initial scheduled execution condition, when an ignitor is triggered one or more tasks will be executed in the cluster. Based on the work-flow and the success condition of these tasks, more tasks could be triggered, and so on.

Other interesting things could happen when a task is terminated (with a success or a failure). You are free to customise this behavior: for instance you could trigger metrics or alerts, or you could program your IoT coffè machine to prepare a good Espresso.

Because Orca is fully integrated with the kubernetes cluster, you are able to run commands via kubectl, exactly as tasks and ignitors were standard Kubernetes resources. Is it not fantastic?

Configurations
---

You do not need to provide complex configurations in order to use Orca. Why does everything always need to be complex? The main concept in orca is that **each task is only responsible to know when it needs to be triggered**. There are only three cases: and ignitor was triggered, another tag succeeded, another tag failed. At the moment complex configurations (eg: AND conditions) are not fully supported. This is in general what you need to know when you design a complex work-flow: in which conditions a specific task should be executed?

Because each task does not need to know the complete work-flow (but just the names of the tasks or the ignitors that trigger it) write configurations is very simple: here few examples.

```
# Task1
---
apiVersion: sirocco.cloud/v1alpha1
kind: Task
metadata:
  name: task1
spec:
  description: A description of the task
  
  startOnIgnition:
    - ignitor1
  startOnSuccess:
  startOnFailure:
    - task1
  
  successActions:
  failureActions:
    - YOUR-PLUGIN-NAME
  
  template:
    spec:
      containers:
      - name: continerName1
        image: dockerImage1
        command: ["whatever command"]
      restartPolicy: Never
```

```
# Task2
---
apiVersion: sirocco.cloud/v1alpha1
kind: Task
metadata:
  name: task2
spec:
  description: A description of the task
  
  startOnIgnition:
  startOnSuccess:
  	- task1
  startOnFailure:
  
  successActions:
  failureActions:
  
  template:
    spec:
      containers:
      - name: continerName2
        image: dockerImage2
        command: ["whatever command"]
      restartPolicy: Never
```

This is a simple example of an ignitor. Note that an ignitor is a kind of scheduler:

```
# Ignitor1
---
apiVersion: sirocco.cloud/v1alpha1
kind: Ignitor
metadata:
  name: ignitor1
spec:
  description: Initiate the workflow ABC
  
  # Note: an ignitor can be scheduled in 3 different ways!
  #scheduled: "2020-02-03T19:34:05Z"
  scheduled: "NOW"
  #scheduled: "5 * * * *"

```

Comunication beteen task
---

Communication between tasks

Sometimes a task needs data from the one initiated it. In Orca this is very simple. Immagine you have two tasks: task1 and task2. The only thing you need to do in task1 is to save the output you want to send to task2 in a file inside a standard Kubernetes location (/dev/termination-log, see http://bit.ly/2vdAFvP). Orca will take care to retrive the content of this file - that is exposed by kubernetes itself, and “inject” inside a special environment variable in the pod initiated by task2. Isn’t it simple?

Orca prefills two special environments: “**ORCA_INITIATOR**” contains the name of the initiator task (or the name of the ignitor), while “**ORCA_DATA**” contains the message string from the previous task (or the content of the field “data” n the ignitor configuration).


Kubectl integrations
---

As said, Tasks and Ignitors are now 100% kubernetes-like object in your cluster, so they can be created, deleted, edited and listed via kubectl (or any other kubernetes-compatible client applications). Here some examples:

```
$ kubectl apply -f task1.yaml 
task.sirocco.cloud/task1 configured
```

```
$ kubectl apply -f ignitor1.yaml 
ignitor.sirocco.cloud/ignitor1 configured
```

```
$kubectl get tasks
NAME    STATE     LASTSUCCESS           LASTFAILURE           FAILURESCOUNT   IGNITORS     ONSUCCESSTASK   ONFAILURETASK
task1   Pending   2020-02-07 16:15:14   2020-02-07 16:13:32   0               [ignitor1]                   [task1]
task2   Pending   2020-02-07 16:15:18                         0                            [task1]         
```

And so on...
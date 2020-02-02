package operator

import (
"fmt"
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"github.com/gorhill/cronexpr"
	"time"
	"strings"
	"errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (o *Operator) registerIgnitors() error{
	// Initialize the initial flow (based on task and ignitors already present in the cluster)
	ignitorsInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Ignitors()
	ignitors, err := ignitorsInformer.Lister().Ignitors("default").List(labels.Everything())
	if err != nil {
		return err
	}

	flow := o.flow
	// For each existing ignitor
	for _, ignitor := range ignitors {
		// Register the ignitor using the name
		name := ignitor.ObjectMeta.Name
		ign, err := flow.RegisterIgnitor(name)
		if err != nil {
			o.log.Error.Println(err)
			continue
		}
		ign.SetGeneration(ignitor.ObjectMeta.Generation)

		// Set the schedule for this ingitor
		if err := o.registerIgnitorSchedule(ignitor); err != nil{
			o.log.Error.Println(err)
			continue
		}
	}

	return nil
}


func (o *Operator) addedIgnitorHandler(new interface{}) {
	if !o.initialized {
		return
	}

	o.log.Trace.Println("Ignitor added event")

	ignitor, ok := new.(*v1alpha1.Ignitor)
	if !ok {
		o.log.Error.Println("object is not an ignitor")
		return
	}

	name := ignitor.ObjectMeta.Name

	// TODO: change in RegisterTaskWithOptions
	if ign, err := o.flow.RegisterIgnitor(name); err != nil {
		o.log.Error.Println("can not register ignitor " + name)
		o.log.Error.Println(err)
		return
	} else {
		ign.SetGeneration(ignitor.ObjectMeta.Generation)

		// Set the schedule for this ingitor
		if err := o.registerIgnitorSchedule(ignitor); err != nil{
			o.log.Error.Println(err)
			return
		}
		o.log.Info.Println("Registered ignitor " + name)
	}

}

func (o *Operator) registerIgnitorSchedule(ignitor *v1alpha1.Ignitor) error{
	name := ignitor.ObjectMeta.Name
	scheduleInterval, err := nextExecutionInterval(ignitor.Spec.Scheduled)
	if err != nil {
		return errors.New("Invalid schedule for ignitor " + name + ". It will be ignored.")			
	} else {
		sec := fmt.Sprintf("%v", scheduleInterval)
		o.log.Trace.Println("Scheduled ignitor " + name + " in " + sec)
		o.ququeIgnitorExecution(scheduleInterval, name)
	}
	return nil
}

func nextExecutionInterval(schedule string) (time.Duration, error){
	// Immediate execution
	if strings.ToUpper(schedule) == "NOW"{
		return 0 * time.Second, nil
	}

	// Cronjob-like schedule
	parsedCronExpression, err := cronexpr.Parse(schedule)
	if err == nil {
		now := time.Now()
		return parsedCronExpression.Next(now).Sub(now), nil
	}

	return 0 * time.Second, errors.New("invalid schedule")
}

func (o *Operator) updatedIgnitorHandler(old, new interface{}) {
	if !o.initialized {
		return
	}

	o.log.Trace.Println("Ignitor updated event")

	ignitor, ok := new.(*v1alpha1.Ignitor)
	if !ok {
		o.log.Error.Println("object is not an ignitor")
		return
	}

	ignitorName := ignitor.ObjectMeta.Name

	ign, err := o.flow.GetIgnitor(ignitorName)
	if err != nil {
		o.log.Error.Println(err)
		return
	}

	if ign.IsUpdated(ignitor.ObjectMeta.Generation) == false {
		o.log.Trace.Println("no changes for " + ignitorName)
		return
	}

	o.log.Trace.Println("Updating " + ignitorName)
}

func (o *Operator) deletedIgnitorHandler(old interface{}) {
	if !o.initialized {
		return
	}
	o.log.Trace.Println("Ignitor deleted event")
	ignitor, ok := old.(*v1alpha1.Ignitor)

	if !ok {
		o.log.Error.Println("object is not an ignitor")
		return
	}

	name := ignitor.ObjectMeta.Name

	if ok := o.flow.RemoveIgnitor(name); ok {
		o.log.Info.Println("Removed ignitor " + name)
	}
}


func (o *Operator) getIgnitorByName(ignitorName string) (*v1alpha1.Ignitor, error) {
	ignitorsInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Ignitors()
	ignitor, err := ignitorsInformer.Lister().Ignitors("default").Get(ignitorName)
	if err != nil {
		return nil, errors.New("ignitor " + ignitorName + " is not regitered in the cluster")
	}
	return ignitor, nil
}

func (o  *Operator) executeIgnitor(ignitorName string, done func()) error {
	// Be sure that the ignitor at this point is still present in the cluster
	ignitor, err := o.getIgnitorByName(ignitorName)
	if err != nil {
		return err
	}

	// In case the ingitor scheduler was "NOW", it need to be removed from the cluster
	if strings.ToUpper(ignitor.Spec.Scheduled) == "NOW"{
		o.ququeIgnitorDeletion(ignitorName)
	} else {
		// Otherwise reschedule
		if err := o.registerIgnitorSchedule(ignitor); err != nil{
			o.log.Error.Println(err)
			o.log.Info.Println("Ignitor " + ignitorName + " can not be rescheduled")			
		}
	}

	o.log.Info.Println("Executing ingitor " + ignitorName)

	// Find the relevant tasks to start
	flowIgnitor, err := o.flow.GetIgnitor(ignitorName)
	if err != nil {
		return err
	}
	for _, taskName := range flowIgnitor.GetTaskNamesToExecute(){
		// Be sure the task is still registered in the internal model
		_, err := o.flow.GetTask(taskName)
		if err != nil {
			continue
		}
		// Add the task in the execution queue
		o.ququeTaskExecution(taskName)
	}

	done()
	return nil
}

func (o  *Operator) deleteIgnitor(ignitorName string, done func()) error {
	// Be sure that the ignitor at this point is still present in the cluster
	_, err := o.getIgnitorByName(ignitorName)
	if err != nil {
		return err
	}

	if err := o.orcaClientSet.SiroccoV1alpha1().Ignitors("default").Delete(ignitorName, &metaV1.DeleteOptions{}); err != nil {
		return err
	}
	
	done()
	return nil
}	
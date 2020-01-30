package operator

import (
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
)

func (o *Operator) addedIgnitorHandler(new interface{}) {
	if !o.initialized {
		return
	}
	o.log.Trace.Println("Ignitor added event")
}


func (o *Operator) updatedIgnitorHandler(old, new interface{}) {
	if !o.initialized {
		return
	}
	o.log.Trace.Println("Task update event")
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
	Show(name)
	if ok := o.flow.RemoveIgnitor(name); ok {
		o.log.Info.Println("Removed ignitor " + name)
	}
}

package operator

import (
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
)

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

	// TODO: change in RegisterTaskWithOptions
	if ign, err := o.flow.RegisterIgnitor(ignitor.ObjectMeta.Name); err != nil {
		o.log.Error.Println("can not register ignitor " + ignitor.ObjectMeta.Name)
		o.log.Error.Println(err)
		return
	} else {
		o.log.Info.Println("Registered ignitor " + ignitor.ObjectMeta.Name)
		ign.SetGeneration(ignitor.ObjectMeta.Generation)
	}

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

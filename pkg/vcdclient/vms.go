/*
   Copyright 2021 VMware, Inc.
   SPDX-License-Identifier: Apache-2.0
*/

package vcdclient

import (
	"fmt"
	"k8s.io/klog"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

const (
	// VCDVMIDPrefix is a prefix added to VM objects by VCD. This needs
	// to be removed for query operations.
	VCDVMIDPrefix = "urn:vcloud:vm:"
)

// GetVMNameInVAppsWithPrefix returns a list of vApp Reference if the vApp prefix matches an existing vApp name.
// If no valid vApp is found, it returns a nil VApp reference list and an error
func GetVMNameInVAppsWithPrefix(vdc *govcd.Vdc, vappPrefix string, vmName string) (*govcd.VM, error) {

	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if strings.HasPrefix(resourceReference.Name, vappPrefix) &&
				resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vApp, err := vdc.GetVAppByHref(resourceReference.HREF)
				if err != nil {
					klog.Errorf("unable to get reference of vApp [%s]: [%v]", resourceReference.Name, err)
					continue
				}
				vm, err := vApp.GetVMByName(vmName, false)
				if err == nil {
					klog.Infof("Found vm [%s] in vApp [%s]", vmName, resourceReference.Name)
					return vm, nil
				} else if err == govcd.ErrorEntityNotFound {
					continue
				} else {
					klog.Errorf("error looking for vm [%s] in vApp [%s]: [%v]", vmName,
						resourceReference.Name, err)
					continue
				}
			}
		}
	}
	return nil, govcd.ErrorEntityNotFound
}

// FindVMByName finds a VM in a vApp using the name. The client is expected to have a valid
// bearer token when this function is called.
func (client *Client) FindVMByName(vmName string) (*govcd.VM, error) {
	if vmName == "" {
		return nil, fmt.Errorf("vmName mandatory for FindVMByName")
	}

	klog.Infof("Trying to find vm [%s] by name in vApp with prefix [%s]", vmName, client.ClusterVAppName)
	vm, err := GetVMNameInVAppsWithPrefix(client.vdc, client.ClusterVAppName, vmName)
	if err != nil {
		return nil, fmt.Errorf("unable to find vm [%s] in vApp [%s]: [%v]", vmName, client.ClusterVAppName, err)
	}

	return vm, nil
}

// GetVMIDInVAppsWithPrefix returns a list of vApp Reference if the vApp prefix matches an existing vApp name.
// If no valid vApp is found, it returns a nil VApp reference list and an error
func GetVMIDInVAppsWithPrefix(vdc *govcd.Vdc, vappPrefix string, vmID string) (*govcd.VM, error) {

	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if strings.HasPrefix(resourceReference.Name, vappPrefix) &&
				resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vApp, err := vdc.GetVAppByHref(resourceReference.HREF)
				if err != nil {
					klog.Errorf("unable to get reference of vApp [%s]: [%v]", resourceReference.Name, err)
					continue
				}
				vm, err := vApp.GetVMById(vmID, false)
				if err == nil {
					klog.Infof("Found vm [%s] in vApp [%s]", vmID, resourceReference.Name)
					return vm, nil
				} else if err == govcd.ErrorEntityNotFound {
					continue
				} else {
					klog.Errorf("error looking for vm [%s] in vApp [%s]: [%v]", vmID,
						resourceReference.Name, err)
					continue
				}
			}
		}
	}
	return nil, govcd.ErrorEntityNotFound
}

// FindVMByUUID finds a VM in a vApp using the UUID. The client is expected to have a valid
// bearer token when this function is called.
func (client *Client) FindVMByUUID(vcdVmUUID string) (*govcd.VM, error) {
	if vcdVmUUID == "" {
		return nil, fmt.Errorf("vmUUID mandatory for FindVMByUUID")
	}

	klog.Infof("Trying to find vm [%s] in vApp [%s] by UUID", vcdVmUUID, client.ClusterVAppName)
	vmUUID := strings.TrimPrefix(vcdVmUUID, VCDVMIDPrefix)

	vm, err := GetVMIDInVAppsWithPrefix(client.vdc, client.ClusterVAppName, vmUUID)
	if err != nil {
		return nil, fmt.Errorf("unable to find vApp [%s] by name: [%v]", client.ClusterVAppName, err)
	}

	return vm, nil
}

// IsVmNotAvailable : In VCD, if the VM is not available, it can be an access error or the VM may not be present.
// Hence we sometimes get an error different from govcd.ErrorEntityNotFound
func (client *Client) IsVmNotAvailable(err error) bool {

	if strings.Contains(err.Error(), "Either you need some or all of the following rights [Base]") &&
		strings.Contains(err.Error(), "to perform operations [VAPP_VM_VIEW]") &&
		strings.Contains(err.Error(), "target entity is invalid") {
		return true
	}

	if strings.Contains(err.Error(), "error refreshing VM: cannot refresh VM, Object is empty") {
		return true
	}

	return false
}

package types

import (
	migtypes "github.com/nebuly-ai/nebulnetes/pkg/gpu/mig/types"
	"github.com/nebuly-ai/nebulnetes/pkg/util"
)

type CreateOperation struct {
	MigProfile migtypes.MigProfile
	Quantity   uint8
}

type DeleteOperation struct {
	MigProfile migtypes.MigProfile
	Resources  []migtypes.MigDeviceResource
	Quantity   uint8
}

type MigConfigPlan struct {
	DeleteOperations []DeleteOperation
	CreateOperations []CreateOperation
}

func NewMigConfigPlan(state MigState, desired GPUSpecAnnotationList) MigConfigPlan {
	plan := MigConfigPlan{}

	// Get resources present in current state which MIG profile is not included in spec
	for migProfile, resourceList := range getResourcesNotIncludedInSpec(state, desired).GroupByMigProfile() {
		op := DeleteOperation{
			MigProfile: migProfile,
			Resources:  resourceList,
			Quantity:   uint8(len(resourceList)), // we want all of these resources to be deleted
		}
		plan.addDeleteOp(op)
	}

	// Compute plan for resources contained in spec annotations
	stateResources := state.Flatten().GroupByMigProfile()
	for migProfile, annotations := range desired.GroupByMigProfile() {
		totalDesiredQuantity := 0
		for _, a := range annotations {
			totalDesiredQuantity += a.Quantity
		}

		actualResources := stateResources[migProfile]
		if actualResources == nil {
			actualResources = make(migtypes.MigDeviceResourceList, 0)
		}

		diff := totalDesiredQuantity - len(actualResources)
		if diff > 0 {
			op := CreateOperation{
				MigProfile: migProfile,
				Quantity:   uint8(diff),
			}
			plan.addCreateOp(op)
		}
		if diff < 0 {
			op := DeleteOperation{
				MigProfile: migProfile,
				Quantity:   uint8(util.Abs(diff)),
				Resources:  actualResources,
			}
			plan.addDeleteOp(op)
		}
	}

	return plan
}

func (p *MigConfigPlan) addDeleteOp(op DeleteOperation) {
	p.DeleteOperations = append(p.DeleteOperations, op)
}

func (p *MigConfigPlan) addCreateOp(op CreateOperation) {
	p.CreateOperations = append(p.CreateOperations, op)
}

func (p *MigConfigPlan) IsEmpty() bool {
	return len(p.DeleteOperations) == 0 && len(p.CreateOperations) == 0
}

func getResourcesNotIncludedInSpec(state MigState, specAnnotations GPUSpecAnnotationList) migtypes.MigDeviceResourceList {
	lookup := specAnnotations.GroupByGpuIndex()

	updatedState := state
	for gpuIndex, annotations := range lookup {
		migProfiles := make([]string, 0)
		for _, a := range annotations {
			migProfiles = append(migProfiles, a.GetMigProfileName())
		}
		updatedState = updatedState.WithoutMigProfiles(gpuIndex, migProfiles)
	}

	return updatedState.Flatten()
}
package provider

import (
	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	artifactoryprovider "github.com/criticalstack/stackapps/controllers/provider/artifactoryprovider"
	awss3provider "github.com/criticalstack/stackapps/controllers/provider/awss3provider"
	vaultprovider "github.com/criticalstack/stackapps/controllers/provider/vaultprovider"
)

type Provider interface {
	Values() (interface{}, error)
}

func New(src *featuresv1alpha1.StackValueSource, path string) Provider {
	switch src.Type {
	case artifactoryprovider.Type:
		return artifactoryprovider.New(src, path)
	case awss3provider.Type:
		return awss3provider.New(src, path)
	case vaultprovider.Type:
		return vaultprovider.New(src, path)
	default:
		return nil
	}
}

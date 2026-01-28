package main

import (
	cfg "github.com/conductorone/baton-aruba-central/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("aruba-central", cfg.Config)
}

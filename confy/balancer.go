package confy

import "github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy/provider"

type providerRebalancer struct {
	providers provider.Provider
	next      int
}

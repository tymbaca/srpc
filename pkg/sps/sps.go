package sps

import "math/big"

type SPS struct {
	privateKey *big.Int
	g          *big.Int
}

func (s *SPS) PublicKey() {
	s.privateKey.Rem()
}

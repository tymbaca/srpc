# sTLS - stupid Transport Layer Security

sTLS is sRPC transport layer providing key exchange and symmetric encryption. 
Currently it uses `crypto/ecdh` Diffie-Hellman key exchange
implementation and `chacha20` (XChaCha20) for encryption.

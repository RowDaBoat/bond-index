Example

Run bitcoind in regtest with:
bitcoind -regtest -txindex

Run ord server in regtest with:
ord --regtest server

Create a wallet on regtest with:
ord --regtest wallet create

Get a receive address on regtest with:
ord --regtest wallet receive

Mine 101 blocks (to unlock the coinbase) with:
bitcoin-cli -regtest generatetoaddress 101 <receive address>

Inscribe on regtest with:
ord --regtest wallet inscribe --fee-rate 1 --file <file>

Mine the inscription:
bitcoin-cli -regtest generatetoaddress 1 <receive address>

By default, browsers don't support compression over HTTP. To test compressed content over HTTP, use the --decompress flag:
ord --regtest server --decompress

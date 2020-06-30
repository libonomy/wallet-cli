# Libonomy Wallet Package For Libo-Nodes
The package is developed for the the developers use only which they can utilize when they start exploring the libosystem.

## Functionality
The wallet implements the 
- Account Creation
- Account Details
- Message Signature
- Hex signature with pvt_key
- Text signature using pvt_key
- Account Reuse

other functionalities will be released soon

## Guideline
Make sure that GO is installed and gopath is set properly
It is recommended that after you clone the repo , you should keep it out of the gopath.
### Build for linux:


```bash
make build-linux
```

## Using with a local node 

1. Make sure that you have configured the local node properly and the url to local server is properly set.

2. Build the wallet using the :

```bash
./cli_wallet_linux_amd64
```

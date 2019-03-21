[TOC]

# Astraia Instructions

a simple and efficient light client implementation.

[![Build Status](https://circleci.com/gh/walterkangluo/lightClient/tree/master.svg?style=shield)](https://circleci.com/gh/walterkangluo/lightClient/tree/master)
[![codecov](https://codecov.io/gh/walterkangluo/lightClient/branch/master/graph/badge.svg)](https://codecov.io/gh/walterkangluo/lightClient)

## Getting started

Running it then should be as simple as:

```
$ make all
```

## Testing

```
$ make test
```

## Command line 

| Command | Describe                                                     |
| :------ | ------------------------------------------------------------ |
| console | The astraia console is an interactive shell for the JavaScript runtime environment which exposes a node admin interface as well as the Ðapp JavaScript API. |
| account | Manage accounts, list all existing accounts, import a private key into a new account, create a new account or update an existing account. |

* console
* account
  * new
  * import
  * update
  * list

### console

Start an interactive JavaScript environment.

```
$astraia console
```

----

### accoount

#### astraia account new

Create a new account

```
$astraia account new

Your new account is locked with a password.Please give a password.Do not forget password.
Passphrase: 
Repeat passphrase: 
Address: {d4e34e3eaeb88249b6912239b40de38606542a97}
```

For non-interactive use you supply a plaintext password file as argument to the `--password` flag. 

```
$astraia account new --password /private/password

Address: {f4412afe07ef9941f3b93fbc8178c6f44f10638c}
```

#### astraia account import

Imports an unencrypted private key from file and creates a new account

```
$astraia account import [options] <keyfile>
```

For example:

```
$astraia account import /priveate/keyfile

Your new account is locked with a password. Please give a password. Do not forget this password.
Passphrase: 
Repeat passphrase: 
Address: {a94f5374fce5edbc8e2a8697c15331677e6ebf0b}
```

#### astraia account update

Update an existing account

```
$astatia account update [options] <address>
```

For example:

```
$astatia account update d7997e13f116893df7a144ff9683e8023d435f70

Unlocking account d7997e13f116893df7a144ff9683e8023d435f70 | Attempt 1/3
Passphrase: 
Please give a new password. Do not forget this password.
Passphrase: 
Repeat passphrase: 
```

#### astraia account list

Print summary of existing accounts

```
$astraia account list

Account #0: {c2a4c39cda11440e4bf9d6323861a5f1b47138a0} ...keyfile path...
Account #1: {d4e34e3eaeb88249b6912239b40de38606542a97} ...keyfile path...
Account #2: {f4412afe07ef9941f3b93fbc8178c6f44f10638c} ...keyfile path...
```

For non-interactive use you specify keystore folder as argument to the `--keystore` flag.

```
$astraia account list --keystore /doc

Account #0: {448b41cc9836761fab62c5c485b65da6836f7f15} ...keyfile path...
```

---

## Console

| Instance | Describe                                                     |
| :------- | ------------------------------------------------------------ |
| personal | The personal API manages private keys in the key store.      |
| eth      | The eth API gives you access to interactive with blockchain. |

personal method

* listAccounts
* lockAccount
* newAccount
* signTransaction
* unlockAccount

eth method

* getBalance
* getTransaction
* getTransactionCount
* newWeb3
* sendRawTransaction

#### personal_listAccounts

Return to the list of accounts in the keystore directory.

**Parameters**

1.keystore `string` required:directory for the keystore  (can be null)

**Returns**

`list`   The list of accounts in the keystore directory.

**Example**

```
>personal.listAccounts()

Account #0: {448b41cc9836761fab62c5c485b65da6836f7f15}  ...keyfile path...
Account #1: {6261b2981595a2a2eefdab121ce4d2e90470c869}  ...keyfile path...
Account #2: {467bf5ec3fdcdf24d22abc4500c17b077e1681c4}  ...keyfile path...
Account #3: {96adfc51a5f07990cea5db3daa47e211a5434100}  ...keyfile path...
Account #4: {3cc9e5fe67debcb293755f5830675a3ecbea35ce}  ...keyfile path...
Account #5: {b0c066aa7f29c34f5ad32f900e2349c9dba9642e}  ...keyfile path...
Account #6: {e26fad38b09db076afdd05bab8df8119c870bbf7}  ...keyfile path...
```

---

#### personal_lockAccount

Return none, lock the account, remove account info about private key

**Parameters**

1. address `string` required: The hexadecimal address of the account.

**Returns**

`none`   If succeeds will return none, otherwise it will return an error message .

**Example**

```
> personal.lockAccount("e26fad38b09db076afdd05bab8df8119c870bbf7")

```

------

#### personal_newAccount

Return  hexadecimal address of the account.

**Parameters**

1.password `string` required: Password for creating a new account.

**Returns**

`Address`   The hexadecimal address of the account .

**Example**

```
>personal.newAccount("123")

Address: {aaacc574e67f1e2a357e05bcebb1aa6596500b47}

```

------

#### personal_signTransaction

Return Rlp-encoded transaction signed by private key.

**Parameters**

1.transaction `string` required: Rlp-encoded transaction signed by private key.

2.password `string` required: Password of the keystore file corresponding to the ‘from’ account

**Returns**

`txEncoded`  Rlp-encoded transaction signed by private key.

**Example**

```
> personal.signTransaction({from: '0xb0c066aa7f29c34f5ad32f900e2349c9dba9642e', to: '0xb0c066aa7f29c34f5ad32f900e2349c9dba9642e', value: web3.toWei(1, "wei"), gas:"123", gasPrice:"123", nonce:"1"}, "123")

0xf877f872017b8094b0c066aa7f29c34f5ad32f900e2349c9dba9642e94b0c066aa7f29c34f5ad32f900e2349c9dba9642e01801ca08165c224ab38cf41325fbae788749f55c875cbb0379b90565f5dfc77c4b8593aa033574624fda3a88a5a976fca71e80e68e9def011e297af85cd743fd05439a6acc0c0c0
```

------

#### personal_unlockAccount

Return none, unlock the account, load account info about private key

**Parameters**

1. address `string` required: The hexadecimal address of the account.
2. password `string` required: Password of the keystore file corresponding to the unlock account

**Returns**

`none`   If succeeds will return none, otherwise it will return an error message .

**Example**

```
> personal.unlockAccount("e26fad38b09db076afdd05bab8df8119c870bbf7", "123")

```

---

#### eth_getBalance

Return the balance of the account of given address.

**Parameters**

1. address `string` required: The hexadecimal address of the account.
2. quantity `string` required: Integer block number, or the string `"latest"`, `"earliest"` or `"pending"`

**Returns**

`balance`  integer of the current balance in wei..

**Example**

```
>eth.getBalance("0xe26fad38b09db076afdd05bab8df8119c870bbf7", "latest")

0
```

---

#### eth_getTransaction

Return the transaction count of the account of given address.

**Parameters**

1. txHash `string` required: The hash of the transaction.

**Returns**

`count`  transaction count of the account 

**Example**

```
>eth.getTransaction("0xef8dadde66af80a228e4899055f2ff202d3e6107904b0f05316ebfcc7a31a850")


```

---

#### eth_getTransactionCount

Return the transaction count of the account of given address.

**Parameters**

1. address `string` required: The hexadecimal address of the account.
2. quantity `string` required: Integer block number, or the string `"latest"`, `"earliest"` or `"pending"`

**Returns**

`count`  Transaction count of the account 

**Example**

```
>eth.getTransaction("0xe26fad38b09db076afdd05bab8df8119c870bbf7", "latest")

0
```

------

#### eth_newWeb3

Return hash of the transaction.

**Parameters**

1. hostname `string` required: rpc ip or hostname of the node.
2. port `string` required: rpc port of the node.

**Returns**

`none`  If succeeds will return none, otherwise it will return an error message .

**Example**

```
>eth.newWeb3("127.0.0.1", "47768")

0xef8dadde66af80a228e4899055f2ff202d3e6107904b0f05316ebfcc7a31a850
```

---

#### eth_sendRawTransaction

Return hash of the transaction.

**Parameters**

1.txBytes `string` required: Rlp-encoded transaction signed by private key.

**Returns**

`TxHash`  Hash of the transaction.

**Example**

```
>eth.sendRawTransaction("0xf877f872017b8094b0c066aa7f29c34f5ad32f900e2349c9dba9642e94b0c066aa7f29c34f5ad32f900e2349c9dba9642e01801ca08165c224ab38cf41325fbae788749f55c875cbb0379b90565f5dfc77c4b8593aa033574624fda3a88a5a976fca71e80e68e9def011e297af85cd743fd05439a6acc0c0c0")

0xef8dadde66af80a228e4899055f2ff202d3e6107904b0f05316ebfcc7a31a850
```

---

## Acount Management

### Creating an account

> Note: Remember yours passwords and keyfiles.Be absolutely sure to have a copy of your keyfile and remember the password for that keyfile, and store them both as securely as possible.There are no escape routes here; lose the keyfile or forget your password and all your ether is gone.

**Using** `astraia account new`

Once you have the astraia installed, creating an account is merely a case of executing the `astraia account new` command in a teminal.

Note that you do not have to run the astraia or sync up with the blockchain to use the `astraia account` command.

```
$astraia account new

Your new account is locked with a password.Please give a password.Do not forget password.
Passphrase: 
Repeat passphrase: 
Address: {d4e34e3eaeb88249b6912239b40de38606542a97}
```

For non-interactive use you supply a plaintext password file as argument to the `--password` flag. 

```
$astraia account new --password /private/password

Address: {f4412afe07ef9941f3b93fbc8178c6f44f10638c}
```

> Note: If you do use the `--password` flag with a password file, make sure the file is not readable or even listable for anyone but you

**Using** `astraia account list`

To list all the accounts with keyfile currently in your `keystore` folder use the `list` subcommand of the `astraia account` command:

```
$astraia account list

Account #0: {c2a4c39cda11440e4bf9d6323861a5f1b47138a0} ...keyfile path...
Account #1: {d4e34e3eaeb88249b6912239b40de38606542a97} ...keyfile path...
Account #2: {f4412afe07ef9941f3b93fbc8178c6f44f10638c} ...keyfile path...
```

For non-interactive use you specify keystore folder as argument to the `--keystore` flag.

```
$astraia account list --keystore /doc

Account #0: {448b41cc9836761fab62c5c485b65da6836f7f15} ...keyfile path...
```

**Using astraia console** 

In order to create a new account using astraia, we must first start astraia in console mode:

```
$astraia console 2 >> log_output

Welcome to the astraia JavaScript console!
modules: personal:1.0
>

```

The console allows you to interact with your local accounts by input commands. For example, try the command to list your accounts:

```
> personal.listAccounts()

Account #0: {448b41cc9836761fab62c5c485b65da6836f7f15} ...keyfile path...
Account #1: {6261b2981595a2a2eefdab121ce4d2e90470c869} ...keyfile path...
Account #2: {467bf5ec3fdcdf24d22abc4500c17b077e1681c4} ...keyfile path...
Account #3: {96adfc51a5f07990cea5db3daa47e211a5434100} ...keyfile path...
Account #4: {3cc9e5fe67debcb293755f5830675a3ecbea35ce} ...keyfile path...
Account #5: {b0c066aa7f29c34f5ad32f900e2349c9dba9642e} ...keyfile path...
Account #6: {e26fad38b09db076afdd05bab8df8119c870bbf7} ...keyfile path...
```

You can also create an account from the console:

```
> personal.newAccount()

Passphrase: 
Repeat passphrase: 
Address: {d4e34e3eaeb88249b6912239b40de38606542a97}
```

### Updating an accounnt

You are able to upgrade your keyfile format and/or upgrade your keyfile password.

**Using astratia**

You can update an existing account on the command line with the `update` subcommand with the account address or index as parameter. Remeber that the account index reflects the order of creation(lexicographic order of keyfile name containing the creation time)

```
$astatia account update d7997e13f116893df7a144ff9683e8023d435f70
```

or

```
$astatia account update 3
```

For example:

```
$astatia account update d7997e13f116893df7a144ff9683e8023d435f70

Unlocking account d7997e13f116893df7a144ff9683e8023d435f70 | Attempt 1/3
Passphrase: 
Please give a new password. Do not forget this password.
Passphrase: 
Repeat passphrase: 
```


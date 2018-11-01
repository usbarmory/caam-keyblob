NXP Cryptographic Acceleration and Assurance Module (CAAM) - Linux driver
=========================================================================

The NXP Cryptographic Acceleration and Assurance Module (CAAM) is a built-in
hardware module that implements secure RAM and a dedicated AES cryptographic
engine for encryption/decryption operations.

A device specific random 256-bit OTPMK key is fused in each SoC at
manufacturing time, this key is unreadable and can only be used by the CAAM for
AES encryption/decryption of user data, through the Secure Non-Volatile Storage
(SNVS) companion block.

This directory contains a Linux kernel driver for the CAAM, with the specific
functionality of encrypting/decrypting a data blob (typically an encryption
key) with the OTPMK made available by the SNVS.

The module allocates character device `/dev/caam_kb` for userspace encryption
and decryption operations.

The kernel driver is a port of the original Freescale one for Linux 3.x with
assorted bugfixes.

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com   | andrej@inversepath.com  

Based on a driver from Freescale Semiconductor, Inc.

Compiling
=========

The following instructions assume compilation on a native armv7 architecture,
when cross compiling adjust `ARCH` and `CROSS_COMPILE` variables accordingly.

```
# the Makefile attempts to locate your Linux kernel source tree, if this fails
# it can be passed with a Makefile variable (e.g. `make KERNEL_SRC=path`)
git clone https://github.com/inversepath/caam_keyblob
cd caam_keyblob
make
make modules_install
```

Once installed the resulting module can be loaded in the traditional manner:

```
modprobe caam_keyblob
```

Operation
=========

**IMPORTANT**: the unique OTPMK internal key is available only when Secure Boot
(HAB) is enabled, otherwise a Non-volatile Test Key (NVTK), identical for each
SoC, is used. The secure operation of the CAAM and SNVS, in production
deployments, should always be paired with Secure Boot activation.

The `caam_keyblob` module, when not in Secure State, issues the following
warning at load time:

```
caam_keyblob: WARNING - not in Secure State, Non-volatile test key in effect
```

When Secure State is correctly detected the module issues following message at
load time:

```
caam_keyblob: Secure State detected
```

The following IOCTL is defined for character device `/dev/caam_kb`:

```
ioctl(file, mode, (caam_kb_data *) kb)
```

The mode can be either `CAAM_KB_ENCRYPT` or `CAAM_KB_DECRYPT` to respectively
select AES-256 encryption or decryption with CBC-MAC (CCM). The mode values and
the `caam_kb_data` structure format are defined in
[caam_keyblob.h](https://github.com/inversepath/caam-keyblob/blob/master/caam_keyblob.h).

The following steps, all taken internally within the CAAM, describe the
encryption operation:

  1. A random 256-bit blob encryption key (DEK) is generated within the CAAM,
     using its internal RNG.

  2. The DEK is used to encrypt the desired data via the CAAM AES-CCM function,
     providing confidentiality and integrity protection.

  3. The DEK is AES-ECB encrypted with a key derived from the OTPMK, using NIST
     SP 800-56A Single-step Key-Derivation Function (5.8.1). An optional key
     modifier can be passed, to be concatenated with the OTPMK and further
     differentiate the key derivation process.

  4. The encrypted DEK and encrypted data are returned in a blob file.

The decryption operation works in reverse, taking the blob file as input and
returning the cleartext data.

The CAAM uses the same nonce and counter block values for every operation,
however each encryption operation uses a different DEK key, generated with the
internal RNG, satisfying the need to avoid key re-use.

The maximum data size is 65487 bytes, this limitation is consistent with the
fact that such data typically consists of an encryption key.

The [INTERLOCK](https://github.com/inversepath/interlock) file encryption
front-end supports the CAAM through this driver, providing a Go userspace
implementation reference.

A standalone Go tool, for encryption and decryption, is also available in the
[caam_keyblob.go](https://github.com/inversepath/caam-keyblob/blob/master/caam_tool.go)
file.

License
=======

NXP Cryptographic Acceleration and Assurance Module (CAAM) - Linux driver
https://github.com/inversepath/caam-keyblob

Copyright (c) F-Secure Corporation  
Copyright (c) 2015 Freescale Semiconductor, Inc.

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.

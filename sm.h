/*
 * CAAM Secure Memory/Keywrap API Definitions
 * Copyright (C) 2008-2015 Freescale Semiconductor, Inc.
 */

#ifndef SM_H
#define SM_H

/* Storage access permissions */
#define SM_PERM_READ 0x01
#define SM_PERM_WRITE 0x02
#define SM_PERM_BLOB 0x03

/* Define treatment of secure memory vs. general memory blobs */
#define SM_SECMEM 0
#define SM_GENMEM 1

/* Define treatment of red/black keys */
#define RED_KEY 0
#define BLACK_KEY 1

/* Define key encryption/covering options */
#define KEY_COVER_ECB 0	/* cover key in AES-ECB */
#define KEY_COVER_CCM 1 /* cover key with AES-CCM */

/*
 * Round a key size up to an AES blocksize boundary so to allow for
 * padding out to a full block
 */
#define AES_BLOCK_PAD(x) ((x % 16) ? ((x >> 4) + 1) << 4 : x)

/* Define space required for BKEK + MAC tag storage in any blob */
#define BLOB_OVERHEAD (32 + 16)

#endif /* SM_H */

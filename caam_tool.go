// NXP Cryptographic Acceleration and Assurance Module (CAAM)
// https://github.com/usbarmory/caam-keyblob
//
// userspace driver reference example
//
// Copyright (c) F-Secure Corporation
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation under version 3 of the License.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// See accompanying LICENSE file for full details.
//
// IMPORTANT: the unique OTPMK internal key is available only when Secure Boot
// (HAB) is enabled, otherwise a Non-volatile Test Key (NVTK), identical for
// each SoC, is used. The secure operation of the CAAM and SNVS, in production
// deployments, should always be paired with Secure Boot activation.
//
// +build linux

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"unsafe"
)

const CAAM_DEV = "/dev/caam_kb"

const (
	KEYMOD_LEN      = 16
	BLOB_OVERHEAD   = 32 + 16
	MAX_KEYBLOB_LEN = 65535
	MAX_RAWKEY_LEN  = MAX_KEYBLOB_LEN - BLOB_OVERHEAD
)

// Portability note: the 0x18 within the two CAAM_KB_* constants and the uint32
// types in caam_kb_data reflect a 32-bit architecture.

const (
	// _IOWR(CAAM_KB_MAGIC, 0, struct caam_kb_data)
	CAAM_KB_ENCRYPT = 0xc0184900
	// _IOWR(CAAM_KB_MAGIC, 1, struct caam_kb_data)
	CAAM_KB_DECRYPT = 0xc0184901
)

// C compatible struct of caam_kb_data from caam_keyblob.h
type caam_kb_data struct {
	Text      *byte
	TextLen   uint32
	Blob      *byte
	BlobLen   uint32
	Keymod    *byte
	KeymodLen uint32
}

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		log.Println("usage: [enc|dec] [cleartext file] [blob file]")
	}
}

func main() {
	var err error
	var mode uint32

	var text []byte
	var blob []byte

	var output *os.File

	flag.Parse()

	if len(flag.Args()) != 3 {
		flag.Usage()
		os.Exit(1)
	}

	op := flag.Arg(0)
	textPath := flag.Arg(1)
	blobPath := flag.Arg(2)

	switch op {
	case "enc":
		mode = CAAM_KB_ENCRYPT
	case "dec":
		mode = CAAM_KB_DECRYPT
	default:
		log.Fatal("caam_tool: error, invalid operation")
	}

	defer func() {
		if err != nil {
			log.Fatalf("caam_tool: error, %v", err)
		}
	}()

	kb := &caam_kb_data{}

	// The key modifier is left empty in this reference example, it is
	// concatenated to the OTPMK to further differentiate derived keys.
	keymod := bytes.Repeat([]byte{0x00}, KEYMOD_LEN)
	kb.Keymod = &keymod[0]
	kb.KeymodLen = KEYMOD_LEN

	switch mode {
	case CAAM_KB_ENCRYPT:
		text, err = ioutil.ReadFile(textPath)

		if err != nil {
			return
		}

		if len(text) > MAX_RAWKEY_LEN {
			log.Fatalf("caam_tool: error, input from %s cannot exceed %d", textPath, MAX_RAWKEY_LEN)
		}

		kb.TextLen = uint32(len(text))
		kb.BlobLen = uint32(len(text) + BLOB_OVERHEAD)
		blob = make([]byte, kb.BlobLen)

		log.Printf("caam_tool: encrypting %d bytes from %s", kb.TextLen, textPath)
	case CAAM_KB_DECRYPT:
		blob, err = ioutil.ReadFile(blobPath)

		if err != nil {
			return
		}

		if len(blob) > MAX_KEYBLOB_LEN {
			log.Fatalf("caam_tool: error, input from %s cannot exceed %d", textPath, MAX_KEYBLOB_LEN)
		}

		kb.BlobLen = uint32(len(blob))
		kb.TextLen = uint32(len(blob) - BLOB_OVERHEAD)
		text = make([]byte, kb.TextLen)

		log.Printf("caam_tool: decrypting %d bytes from %s", kb.BlobLen, blobPath)
	}

	kb.Text = &text[0]
	kb.Blob = &blob[0]

	caam, err := os.OpenFile(CAAM_DEV, os.O_RDWR, 0600)

	if err != nil {
		return
	}

	syscall.Flock(int(caam.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(caam.Fd()), syscall.LOCK_UN)
	defer caam.Close()

	log.Printf("caam_tool: caam_kb_data %+v", kb)
	log.Printf("caam_tool: issuing ioctl %x on %s", mode, CAAM_DEV)

	err = ioctl(caam.Fd(), uintptr(mode), uintptr(unsafe.Pointer(kb)))

	if err != nil {
		return
	}
	defer caam.Close()

	switch mode {
	case CAAM_KB_ENCRYPT:
		output, err = os.OpenFile(blobPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_EXCL|os.O_SYNC, 0600)

		if err != nil {
			return
		}

		output.Write(blob)

		log.Printf("caam_tool: encrypted %d bytes to %s", kb.BlobLen, blobPath)
	case CAAM_KB_DECRYPT:
		output, err = os.OpenFile(textPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_EXCL|os.O_SYNC, 0600)

		if err != nil {
			return
		}

		output.Write(text)

		log.Printf("caam_tool: decrypted %d bytes to %s", kb.TextLen, textPath)
	}

	output.Close()
}

func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, arg)

	if e != 0 {
		return syscall.Errno(e)
	}

	return
}

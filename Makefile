ifneq ($(KERNELRELEASE),)

# kbuild
ccflags-y := -march=armv7-a
ccflags-y += -I$(srctree)/drivers/crypto/caam
obj-m += caam_keyblob.o

else

KERNEL_SRC ?= /lib/modules/$(shell uname -r)/build
GO ?= go

.PHONY: caam_tool

all:
	make -C ${KERNEL_SRC} M=$(CURDIR) modules

modules_install:
	make -C ${KERNEL_SRC} M=$(CURDIR) modules_install

clean:
	make -C ${KERNEL_SRC} M=$(CURDIR) clean

caam_tool:
	GOARCH=arm ${GO} build -ldflags "-s -w" -o caam_tool caam_tool.go
endif

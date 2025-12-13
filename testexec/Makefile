# testexec/Makefile - Lima VM management for integration tests

LIMA_INSTANCE := mynt
LIMA_CONFIG := lima-mynt.yaml

.PHONY: help vm-create vm-start vm-stop vm-delete vm-shell vm-status test-integration

help:
	@echo "Lima VM Test Environment"
	@echo ""
	@echo "VM Management:"
	@echo "  make vm-create   - Create and start the Lima VM"
	@echo "  make vm-start    - Start an existing VM"
	@echo "  make vm-stop     - Stop the VM"
	@echo "  make vm-delete   - Delete the VM completely"
	@echo "  make vm-shell    - Open a shell in the VM"
	@echo "  make vm-status   - Show VM status and ZFS version"
	@echo ""
	@echo "Testing:"
	@echo "  make test-integration - Run integration tests on VM"
	@echo ""
	@echo "Environment Variables:"
	@echo "  LIMA_INSTANCE - VM instance name (default: mynt)"

# Create and start VM
vm-create:
	limactl create --name=$(LIMA_INSTANCE) $(LIMA_CONFIG)
	limactl start $(LIMA_INSTANCE)
	@echo ""
	@echo "VM created. Waiting for ZFS provisioning..."
	@sleep 5
	limactl shell $(LIMA_INSTANCE) zfs version

vm-start:
	limactl start $(LIMA_INSTANCE)

vm-stop:
	limactl stop $(LIMA_INSTANCE)

vm-delete:
	limactl delete -f $(LIMA_INSTANCE)

vm-shell:
	limactl shell $(LIMA_INSTANCE)

vm-status:
	@limactl list | grep $(LIMA_INSTANCE) || echo "VM not found"
	@echo ""
	@limactl shell $(LIMA_INSTANCE) zfs version 2>/dev/null || echo "ZFS not available"

# Run integration tests on the VM
# Cross-compile for Linux, execute via lima-exec.sh
test-integration:
	cd .. && GOOS=linux GOARCH=$$(limactl info $(LIMA_INSTANCE) | grep -o '"arch": "[^"]*"' | cut -d'"' -f4) \
		go test -tags=integration -exec ./testexec/lima-exec.sh ./...

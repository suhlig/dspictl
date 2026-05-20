package dspi

// USBControlTransfer abstracts USB vendor-specific control transfers.
type USBControlTransfer interface {
	// ControlTransfer performs a USB control transfer.
	// bmRequestType is the USB request type bitmask (direction, type, recipient).
	// bRequest is the vendor-specific request code.
	// wValue and wIndex are USB setup packet fields whose meaning depends on bRequest.
	// data is the transfer buffer; for IN transfers it receives data, for OUT transfers it
	// holds the payload to send. The returned int is the number of bytes transferred.
	ControlTransfer(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error)

	// Close releases the underlying USB device and any associated resources.
	Close() error
}

// NewDevice creates a Device from a USBControlTransfer and metadata.
func NewDevice(usb USBControlTransfer, platform Platform, serial string, bus, address int) *Device {
	return &Device{
		usb:      usb,
		platform: platform,
		serial:   serial,
		bus:      bus,
		address:  address,
	}
}

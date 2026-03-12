//go:build windows

package main

import (
	"os"
	"golang.org/x/sys/windows"
)

func getOwnerInfo(info os.FileInfo, path string) string {
	// Get the security descriptor for the file
	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	if err != nil {
		return "Unknown-User"
	}

	// Extract the SID of the owner
	ownerSid, _, err := sd.Owner()
	if err != nil {
		return "Unknown-SID"
	}

	// Convert SID to a readable name (e.g., "DESKTOP-ABC\User")
	name, domain, _, err := ownerSid.LookupAccount("")
	if err != nil {
		return ownerSid.String() // Fallback to SID string if name lookup fails
	}

	return domain + "\\" + name
}
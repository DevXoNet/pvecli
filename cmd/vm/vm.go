// Copyright 2025 DevXo part of vByte Ltd //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vm

import "github.com/spf13/cobra"

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "VM-specific operations",
	Long:  "Commands that only work with QEMU VMs (not LXC containers)",
}

// Init registers all VM-specific commands
func Init(root *cobra.Command) {
	vmCmd.AddCommand(suspendCmd)
	vmCmd.AddCommand(resumeCmd)
	vmCmd.AddCommand(templateCmd)
	vmCmd.AddCommand(diskCmd)
	vmCmd.AddCommand(cloneCmd)
	vmCmd.AddCommand(migrateCmd)

	root.AddCommand(vmCmd)
}

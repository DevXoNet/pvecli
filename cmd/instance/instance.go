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

package instance

import "github.com/spf13/cobra"

// Init registers all instance commands (shared between VM and LXC)
func Init(root *cobra.Command) {
	root.AddCommand(startCmd)
	root.AddCommand(stopCmd)
	root.AddCommand(rebootCmd)
	root.AddCommand(shutdownCmd)
	root.AddCommand(statusCmd)
	root.AddCommand(infoCmd)
	root.AddCommand(consoleCmd)
	root.AddCommand(snapshotCmd)
	// novncCmd removed - functionality merged into consoleCmd
}

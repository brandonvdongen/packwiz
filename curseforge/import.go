package curseforge

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/comp500/packwiz/core"
	"github.com/spf13/cobra"
)

type twitchPackMeta struct {
	Name string `json:"name"`
	Path string `json:"installPath"`
	// TODO: javaArgsOverride?
	// TODO: allocatedMemory?
	MCVersion string `json:"gameVersion"`
	Modloader struct {
		Name string `json:"name"`
	} `json:"baseModLoader"`
	// TODO: modpackOverrides?
	Mods []struct {
		ID   int `json:"addonID"`
		File struct {
			// This is exactly the same as modFileInfo, but with capitalised
			// FileNameOnDisk.
			ID           int          `json:"id"`
			FileName     string       `json:"FileNameOnDisk"`
			FriendlyName string       `json:"fileName"`
			Date         cfDateFormat `json:"fileDate"`
			Length       int          `json:"fileLength"`
			FileType     int          `json:"releaseType"`
			// fileStatus? means latest/preferred?
			DownloadURL  string   `json:"downloadUrl"`
			GameVersions []string `json:"gameVersion"`
			Fingerprint  int      `json:"packageFingerprint"`
			Dependencies []struct {
				ModID int `json:"addonId"`
				Type  int `json:"type"`
			} `json:"dependencies"`
		} `json:"installedFile"`
	} `json:"installedAddons"`
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an installed curseforge modpack",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pack, err := core.LoadPack()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		index, err := pack.LoadIndex()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var packMeta twitchPackMeta
		// TODO: is this relative to something?
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = json.NewDecoder(f).Decode(&packMeta)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		modIDs := make([]int, len(packMeta.Mods))
		for i, v := range packMeta.Mods {
			modIDs[i] = v.ID
		}

		fmt.Println("Querying Curse API...")

		modInfos, err := getModInfoMultiple(modIDs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		modInfosMap := make(map[int]modInfo)
		for _, v := range modInfos {
			modInfosMap[v.ID] = v
		}

		// TODO: multithreading????
		for _, v := range packMeta.Mods {
			modInfoValue, ok := modInfosMap[v.ID]
			if !ok {
				if len(v.File.FriendlyName) > 0 {
					fmt.Printf("Failed to obtain mod information for \"%s\"\n", v.File.FriendlyName)
				} else {
					fmt.Printf("Failed to obtain mod information for \"%s\"\n", v.File.FileName)
				}
				continue
			}

			fmt.Printf("Imported mod \"%s\" successfully!\n", modInfoValue.Name)

			err = createModFile(modInfoValue, modFileInfo(v.File), &index)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		// TODO: import existing files (config etc.)

		err = index.Write()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = pack.UpdateIndexHash()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = pack.Write()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	curseforgeCmd.AddCommand(importCmd)
}

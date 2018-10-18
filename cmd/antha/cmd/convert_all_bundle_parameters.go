package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertAllBundlesCmd = &cobra.Command{
	Use:   "all-bundle-parameters",
	Short: "update all bundle parameters according to the map of old parameter names to new found in metadata",
	RunE:  convertAllBundles,
}

type bundle struct {
	Dir      string
	Path     string
	FileName string
}

type bundles struct {
	Bundles []*bundle
	seen    map[string]bool
}

func newBundles() *bundles {
	return &bundles{
		seen: make(map[string]bool),
	}
}

func (b *bundles) Walk(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}
	if !strings.HasSuffix(path, ".bundle.json") {
		return nil
	}

	dir, fileName := filepath.Split(path)
	if b.seen[dir] {
		return nil
	}

	b.seen[dir] = true

	b.Bundles = append(b.Bundles, &bundle{
		Dir:      dir,
		Path:     path,
		FileName: fileName,
	})

	return nil
}

//
func convertAllBundles(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// find metadata files
	elements := newElements()
	if err := filepath.Walk(viper.GetString("rootDir"), elements.Walk); err != nil {
		return err
	}

	bundles := newBundles()

	if path := viper.GetString("specificFile"); path != "" {

		dir, fileName := filepath.Split(path)

		bundles.seen[dir] = true

		bundles.Bundles = append(bundles.Bundles, &bundle{
			Path:     path,
			Dir:      dir,
			FileName: fileName,
		})
	} else if err := filepath.Walk(viper.GetString("rootDir"), bundles.Walk); err != nil {
		return err
	}

	var errs []string

	for _, elem := range elements.Elements {

		metadataFileName := filepath.Join(elem.Dir, "metadata.json")

		cFile, err := os.Open(metadataFileName)

		if err != nil {
			errs = append(errs, metadataFileName+": ", err.Error())
			cFile.Close() //nolint
		} else {
			var c NewElementMappingDetails
			decConv := json.NewDecoder(cFile)
			if err := decConv.Decode(&c); err != nil {
				errs = append(errs, "error decoding to NewElementMappingDetails for "+metadataFileName+": ", err.Error())
			}
			cFile.Close() //nolint

			if !c.Empty() {
				for i, bundle := range bundles.Bundles {

					fileName := bundle.FileName

					if !strings.HasPrefix(bundle.FileName, viper.GetString("addPrefix")) {
						fileName = viper.GetString("addPrefix") + bundle.FileName
					}

					err := convertBundleWithArgs(metadataFileName, bundle.Path, filepath.Join(bundle.Dir, fileName))
					if err != nil {
						errs = append(errs, metadataFileName+" + "+bundle.Path+": "+err.Error())
					}
					// update bundle name, in case it will be re-modified
					bundles.Bundles[i].FileName = fileName
				}
			}
		}
	}

	if len(errs) > 0 {
		return errors.Errorf(strings.Join(errs, "\n"))
	}

	return nil
}

func init() {
	c := convertAllBundlesCmd
	convertCmd.AddCommand(c)

	flags := c.Flags()
	flags.String("rootDir", ".", "root directory to search for metadata files with new element mapping and test bundles to update")
	flags.String("addPrefix", "", "adds a common prefix to the start of all updated bundle files")
	flags.String("specificFile", "", "specify a single bundle file to convert with all metadata files found in rootDir")
}

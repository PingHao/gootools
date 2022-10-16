/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("prune called")
		scan(args)
		conclude()
	},
}

type lstfile map[string]struct{}

var (
	log  *zap.Logger
	dict = map[string]lstfile{}
)

func addFile(name string) {
	stat, err := os.Stat(name)
	if err != nil || stat.IsDir() {
		return
	}
	b, err := cksumSha1(name)
	if err != nil {
		return
	}
	c := base64.URLEncoding.EncodeToString(b)
	lf, ok := dict[c]
	if !ok {
		lf = make(lstfile)
	}
	lf[name] = struct{}{}
	dict[c] = lf
}

func init() {
	rootCmd.AddCommand(pruneCmd)
	log, _ = zap.NewDevelopment()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pruneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pruneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func scan(dirs []string) {
	log.Info("run", zap.Any("directories", dirs))
	seen := make(map[string]struct{})
	for _, d := range dirs {

		d, err := filepath.Abs(d)
		if err != nil {
			log.Error("fail to get abs ", zap.Error(err))
			continue
		}
		if _, ok := seen[d]; ok {
			log.Info("already seen " + d)
		}
		seen[d] = struct{}{}
		stat, err := os.Stat(d)
		if err != nil {
			log.Error("fail to get stat "+d, zap.Error(err))
			continue
		}
		if stat.IsDir() {
			proc(d)
		} else {
			log.Info("not directory " + d)
		}
	}
}

func conclude() {
	log.Info("", zap.Int("size of dict", len(dict)))
	for sum, files := range dict {
		if len(files) > 1 {
			log.Info("found duplication", zap.String("sha1", sum), zap.Any("files", files))
		}
	}
}

func proc(dir string) {
	log.Info("start proc directory", zap.String("directory", dir))
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			addFile(path)
			return nil
		})
	if err != nil {
		log.Error("error", zap.Error(err))
	}
}

func cksumSha1(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

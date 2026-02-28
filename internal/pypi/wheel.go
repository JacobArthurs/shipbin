package pypi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/jacobarthurs/shipbin/internal/config"
)

var pep440Re = regexp.MustCompile(`^\d+(\.\d+)*((a|b|rc)\d+)?(\.post\d+)?(\.dev\d+)?$`)

func toPyPIVersion(v string) (string, error) {
	if !pep440Re.MatchString(v) {
		return "", fmt.Errorf(
			"version %q is not valid PEP 440 (required by PyPI)\n"+
				"examples: 1.0.0, 1.0.0a1, 1.0.0b1, 1.0.0rc1, 1.0.0.dev1",
			v,
		)
	}
	return v, nil
}

type wheelFile struct {
	filename        string
	pkgName         string
	version         string
	summary         string
	license         string
	description     string
	descContentType string
	data            []byte
}

func buildWheel(cfg *Config, a config.Artifact) (wheelFile, error) {
	name := strings.NewReplacer("-", "_", ".", "_").Replace(cfg.Name)
	version, err := toPyPIVersion(cfg.Version)
	if err != nil {
		return wheelFile{}, err
	}
	wheelTag := a.Mapping.PyPI.WheelTag
	filename := fmt.Sprintf("%s-%s-py3-none-%s.whl", name, version, wheelTag)
	distInfo := fmt.Sprintf("%s-%s.dist-info", name, version)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	record := &strings.Builder{}

	binaryName := cfg.Name
	if a.Platform.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := fmt.Sprintf("%s/bin/%s", name, binaryName)
	binaryData, err := os.ReadFile(a.Path)
	if err != nil {
		return wheelFile{}, fmt.Errorf("failed to read binary %s: %w", a.Path, err)
	}
	if err := addFileToZip(zw, binaryPath, binaryData, 0755, record); err != nil {
		return wheelFile{}, err
	}

	shimData, err := renderShim(cfg.Name)
	if err != nil {
		return wheelFile{}, fmt.Errorf("failed to render shim: %w", err)
	}
	initPath := fmt.Sprintf("%s/__init__.py", name)
	if err := addFileToZip(zw, initPath, shimData, 0644, record); err != nil {
		return wheelFile{}, err
	}

	readmeContent, contentType, err := readReadme(cfg.Readme)
	if err != nil {
		return wheelFile{}, fmt.Errorf("failed to read readme: %w", err)
	}
	metadata := buildMetadata(name, version, cfg.Summary, cfg.License, readmeContent, contentType)
	metadataPath := fmt.Sprintf("%s/METADATA", distInfo)
	if err := addFileToZip(zw, metadataPath, []byte(metadata), 0644, record); err != nil {
		return wheelFile{}, err
	}

	wheelMeta := buildWheelMeta(wheelTag)
	wheelMetaPath := fmt.Sprintf("%s/WHEEL", distInfo)
	if err := addFileToZip(zw, wheelMetaPath, []byte(wheelMeta), 0644, record); err != nil {
		return wheelFile{}, err
	}

	entryPoints := fmt.Sprintf("[console_scripts]\n%s = %s:main\n", cfg.Name, name)
	entryPointsPath := fmt.Sprintf("%s/entry_points.txt", distInfo)
	if err := addFileToZip(zw, entryPointsPath, []byte(entryPoints), 0644, record); err != nil {
		return wheelFile{}, err
	}

	recordPath := fmt.Sprintf("%s/RECORD", distInfo)
	fmt.Fprintf(record, "%s,,\n", recordPath)
	if err := addFileToZip(zw, recordPath, []byte(record.String()), 0644, nil); err != nil {
		return wheelFile{}, err
	}

	if err := zw.Close(); err != nil {
		return wheelFile{}, fmt.Errorf("failed to close wheel zip: %w", err)
	}

	return wheelFile{
		filename:        filename,
		pkgName:         name,
		version:         version,
		summary:         cfg.Summary,
		license:         cfg.License,
		description:     readmeContent,
		descContentType: contentType,
		data:            buf.Bytes(),
	}, nil
}

func addFileToZip(zw *zip.Writer, path string, data []byte, mode os.FileMode, record *strings.Builder) error {
	header := &zip.FileHeader{
		Name:               path,
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(data),
		CompressedSize64:   uint64(len(data)),
		UncompressedSize64: uint64(len(data)),
	}
	header.SetMode(mode)
	header.ExternalAttrs = uint32(mode) << 16

	w, err := zw.CreateRaw(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry %s: %w", path, err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write zip entry %s: %w", path, err)
	}

	if record != nil {
		hash := sha256.Sum256(data)
		encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
		fmt.Fprintf(record, "%s,sha256=%s,%d\n", path, encoded, len(data))
	}

	return nil
}

func buildMetadata(name, version, summary, license, readmeContent, contentType string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Metadata-Version: 2.1\n")
	fmt.Fprintf(&sb, "Name: %s\n", name)
	fmt.Fprintf(&sb, "Version: %s\n", version)
	if summary != "" {
		fmt.Fprintf(&sb, "Summary: %s\n", summary)
	}
	if license != "" {
		fmt.Fprintf(&sb, "License: %s\n", license)
	}
	fmt.Fprintf(&sb, "Requires-Python: >=3.7\n")
	if readmeContent != "" {
		fmt.Fprintf(&sb, "Description-Content-Type: %s\n", contentType)
		fmt.Fprintf(&sb, "\n%s", readmeContent)
	}
	return sb.String()
}

func readReadme(path string) (content, contentType string, err error) {
	if path == "" {
		return "", "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		contentType = "text/markdown"
	case strings.HasSuffix(lower, ".rst"):
		contentType = "text/x-rst"
	default:
		contentType = "text/plain"
	}
	return string(data), contentType, nil
}

func buildWheelMeta(platformTag string) string {
	return fmt.Sprintf(`Wheel-Version: 1.0
Generator: shipbin
Root-Is-Purelib: false
Tag: py3-none-%s
`, platformTag)
}

func (w wheelFile) reader() io.Reader {
	return bytes.NewReader(w.data)
}

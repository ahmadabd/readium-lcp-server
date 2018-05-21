/*
 * Copyright (c) 2016-2018 Readium Foundation
 *
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 *  1. Redistributions of source code must retain the above copyright notice, this
 *     list of conditions and the following disclaimer.
 *  2. Redistributions in binary form must reproduce the above copyright notice,
 *     this list of conditions and the following disclaimer in the documentation and/or
 *     other materials provided with the distribution.
 *  3. Neither the name of the organization nor the names of its contributors may be
 *     used to endorse or promote products derived from this software without specific
 *     prior written permission
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 *  ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 *  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 *  DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
 *  ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 *  (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 *  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 *  ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 *  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 *  SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package epub

import (
	"archive/zip"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func (ep *Epub) addCleartextResources(names []string) {
	if ep.cleartextResources == nil {
		ep.cleartextResources = []string{}
	}

	for _, name := range names {
		ep.cleartextResources = append(ep.cleartextResources, name)
	}
}

func (ep *Epub) addCleartextResource(name string) {
	if ep.cleartextResources == nil {
		ep.cleartextResources = []string{}
	}

	ep.cleartextResources = append(ep.cleartextResources, name)
}

func (ep Epub) Write(dst io.Writer) error {
	w := NewWriter(dst)

	err := w.WriteHeader()
	if err != nil {
		return err
	}

	for _, res := range ep.Resource {
		if res.Path != "mimetype" {
			fw, err := w.AddResource(res.Path, res.StorageMethod)
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, res.Contents)
			if err != nil {
				return err
			}
		}
	}

	if ep.Encryption != nil {
		writeEncryption(ep, w)
	}

	return w.Close()
}

func (ep Epub) Cover() (bool, *Resource) {

	for _, p := range ep.Package {

		var coverImageID string
		coverImageID = "cover-image"
		for _, meta := range p.Metadata.Metas {
			if meta.Name == "cover" {
				coverImageID = meta.Content
			}
		}

		for _, it := range p.Manifest.Items {

			if strings.Contains(it.Properties, "cover-image") ||
				it.Id == coverImageID {

				path := filepath.Join(p.BasePath, it.Href)
				for _, r := range ep.Resource {
					if r.Path == path {
						return true, r
					}
				}
			}
		}
	}

	return false, nil
}

func (ep *Epub) Add(name string, body io.Reader, size uint64) error {
	ep.Resource = append(ep.Resource, &Resource{Contents: body, StorageMethod: zip.Deflate, Path: name, OriginalSize: size})

	return nil
}

func (ep Epub) CanEncrypt(file string) bool {
	i := sort.SearchStrings(ep.cleartextResources, file)
	return i >= len(ep.cleartextResources) || ep.cleartextResources[i] != file
}

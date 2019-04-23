// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package osutil_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/testutil"
)

type EnsureTreeStateSuite struct {
	dir   string
	globs []string
}

var _ = Suite(&EnsureTreeStateSuite{globs: []string{"*.snap"}})

func (s *EnsureTreeStateSuite) SetUpTest(c *C) {
	s.dir = c.MkDir()
}

func (s *EnsureTreeStateSuite) TestVerifiesExpectedFiles(c *C) {
	c.Assert(os.MkdirAll(filepath.Join(s.dir, "foo", "bar"), 0755), IsNil)
	name := filepath.Join(s.dir, "foo", "bar", "expected.snap")
	c.Assert(ioutil.WriteFile(name, []byte("expected"), 0600), IsNil)
	changed, removed, err := osutil.EnsureTreeState(s.dir, s.globs, map[string]map[string]*osutil.FileState{
		"foo/bar": {
			"expected.snap": {Content: []byte("expected"), Mode: 0600},
		},
	})
	c.Assert(err, IsNil)
	c.Check(changed, HasLen, 0)
	c.Check(removed, HasLen, 0)

	// The content and permissions are correct
	c.Check(name, testutil.FileEquals, "expected")
	stat, err := os.Stat(name)
	c.Assert(err, IsNil)
	c.Check(stat.Mode().Perm(), Equals, os.FileMode(0600))
}

func (s *EnsureTreeStateSuite) TestCreatesMissingFiles(c *C) {
	c.Assert(os.MkdirAll(filepath.Join(s.dir, "foo"), 0755), IsNil)

	changed, removed, err := osutil.EnsureTreeState(s.dir, s.globs, map[string]map[string]*osutil.FileState{
		"foo": {
			"missing1.snap": {Content: []byte(`content-1`), Mode: 0600},
		},
		"bar": {
			"missing2.snap": {Content: []byte(`content-2`), Mode: 0600},
		},
	})
	c.Assert(err, IsNil)
	c.Check(changed, DeepEquals, []string{"bar/missing2.snap", "foo/missing1.snap"})
	c.Check(removed, HasLen, 0)
}

func (s *EnsureTreeStateSuite) TestRemovesUnexpectedFiles(c *C) {
	c.Assert(os.MkdirAll(filepath.Join(s.dir, "foo"), 0755), IsNil)
	c.Assert(os.MkdirAll(filepath.Join(s.dir, "bar"), 0755), IsNil)
	name1 := filepath.Join(s.dir, "foo", "evil1.snap")
	name2 := filepath.Join(s.dir, "bar", "evil2.snap")
	c.Assert(ioutil.WriteFile(name1, []byte(`evil-1`), 0600), IsNil)
	c.Assert(ioutil.WriteFile(name2, []byte(`evil-2`), 0600), IsNil)

	changed, removed, err := osutil.EnsureTreeState(s.dir, s.globs, map[string]map[string]*osutil.FileState{
		"foo": {},
	})
	c.Assert(err, IsNil)
	c.Check(changed, HasLen, 0)
	c.Check(removed, DeepEquals, []string{"bar/evil2.snap", "foo/evil1.snap"})
	c.Check(name1, testutil.FileAbsent)
	c.Check(name2, testutil.FileAbsent)
}

func (s *EnsureTreeStateSuite) TestIgnoresUnrelatedFiles(c *C) {
	c.Assert(os.MkdirAll(filepath.Join(s.dir, "foo"), 0755), IsNil)
	name := filepath.Join(s.dir, "foo", "unrelated")
	err := ioutil.WriteFile(name, []byte(`text`), 0600)
	c.Assert(err, IsNil)
	changed, removed, err := osutil.EnsureTreeState(s.dir, s.globs, map[string]map[string]*osutil.FileState{})
	c.Assert(err, IsNil)
	// Report says that nothing has changed
	c.Check(changed, HasLen, 0)
	c.Check(removed, HasLen, 0)
	// The file is still there
	c.Check(name, testutil.FilePresent)
}

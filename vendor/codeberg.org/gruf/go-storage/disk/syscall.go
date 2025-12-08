package disk

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-fastpath/v2"
	"codeberg.org/gruf/go-storage/internal"
)

// open file for read args.
var readArgs = OpenArgs{
	Flags: syscall.O_RDONLY,
	Perms: 0,
}

// Dirent is a sycall.Dirent wrapper
// with some useful utility methods.
type Dirent struct{ syscall.Dirent }

// DirEntry ...
type DirEntry struct {
	Path string
	Type uint8
}

// IsRegular returns whether entry file type is regular.
func (d *Dirent) IsRegular() bool {
	return d.Dirent.Type == syscall.DT_REG
}

// IsDir returns whether entry file type is directory.
func (d *Dirent) IsDir() bool {
	return d.Dirent.Type == syscall.DT_DIR
}

// IsCurDir returns whether Dirent represents ".".
func (d *Dirent) IsCurDir() bool {
	return d.Name[0] == '.' &&
		d.Name[1] == 0
}

// IsPrevDir returns whether Dirent represents "..".
func (d *Dirent) IsPrevDir() bool {
	return d.Name[0] == '.' &&
		d.Name[1] == '.' &&
		d.Name[2] == 0
}

// NameStr returns the Dirent name as usable Go string.
func (d *Dirent) NameStr() string {

	// Return string copy of name.
	return string(d.nameptr())
}

// nameptr returns a byte slice of the
// currently set Dirent{} name byte slice.
//
// NOTE: IS DIRECTLY TIED TO DIRENT MEMORY!
func (d *Dirent) nameptr() []byte {
	var i int

	// Get string NUL terminator.
	for ; i < len(d.Name); i++ {
		if d.Name[i] == 0 {
			break
		}
	}

	// Empty str.
	if i == 0 {
		return nil
	}

	// Get name slice.
	name := d.Name[:i]

	// Cast []int8 name to string-able []uint8.
	bname := *(*[]byte)(unsafe.Pointer(&name))
	return bname
}

// walk_dir traverses the dir tree of the supplied path, performing the supplied walkFn on each entry.
func walk_dir(pb *fastpath.Builder, path string, walk func(absdir, reldir string, ent *Dirent) error) error {
	type dirframe struct {
		abs string
		rel string
	}

	if walk == nil {
		panic("nil func")
	}

	// ...
	stack := make([]dirframe, 0, 64)
	stack = append(stack, dirframe{
		abs: path,
		rel: "",
	})

	for len(stack) > 0 {
		// Pop next frame from stack.
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Iterate the entries in this frame's directory path.
		if err := readdir(frame.abs, func(ent *Dirent) error {

			if ent.IsDir() {
				// Get a temp. copy of entry name.
				name := byteutil.B2S(ent.nameptr())

				// Append subdirectory entries
				// as frames to walking stack.
				stack = append(stack, dirframe{
					abs: pb.Join(frame.abs, name),
					rel: pb.Join(frame.rel, name),
				})
			}

			// Pass each entry to walk funtion.
			return walk(frame.abs, frame.rel, ent)
		}); err != nil {
			return err
		}
	}

	return nil
}

// clean_dirs traverses the dir tree of supplied
// path, removing any folders with zero children.
func clean_dirs(path string) error {
	pb := internal.GetPathBuilder()
	_, err := clean_dir(pb, path)
	internal.PutPathBuilder(pb)
	return err
}

// clean_dir performs the actual dir cleaning logic for the above top-level version.
func clean_dir(pb *fastpath.Builder, path string) (empty bool, err error) {
	var dirs []string
	empty = true

	// Iterate through entries, collecting subdirs.
	if err = readdir(path, func(ent *Dirent) error {

		// Any entry indicates
		// current is not empty.
		empty = false

		if ent.IsDir() {
			// Get a temp. copy of entry name.
			name := byteutil.B2S(ent.nameptr())

			// Append path of subdir.
			dir := pb.Join(path, name)
			dirs = append(dirs, dir)
		}

		return nil
	}); err != nil {
		return
	}

	if empty {
		return
	}

	// Reset empty.
	empty = true

	var errs []error
	for _, dir := range dirs {
		// Recursively clean subdirectories.
		eachEmpty, err := clean_dir(pb, dir)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if eachEmpty {
			// Dir is empty,
			// now remove it!
			err = rmdir(dir)
			if err != nil {
				err = fmt.Errorf("error deleting subdir %s: %w", dir, err)
				errs = append(errs, err)
				continue
			}
		} else {
			// Unset empty.
			empty = false
		}
	}

	// Return combined errors.
	err = errors.Join(errs...)
	return
}

// readdir is a wrapper to repeatedly call syscall.ReadDirent() on the
// opened file descriptor at path, passing each syscall.Dirent{} to func.
func readdir(path string, each func(*Dirent) error) error {
	if each == nil {
		panic("nil func")
	}

	// Open directory at path for read.
	fd, err := open_fd(path, readArgs)
	if err != nil {
		return err
	}

	// Ensure closed.
	defer close(fd)

	// Alloc default buf size.
	buf := make([]byte, 4096)

	for {
		var nn int

		// Read next block of directory entries into buf.
		if nn, err = readdirent(fd, buf); err != nil {
			return err
		}

		if nn == 0 {
			// Reached end.
			break
		}

	iter:
		for off := 0; off < nn; {
			// Directly cast bytes at offset to Dirent{}.
			ent := (*Dirent)(unsafe.Pointer(&buf[off]))

			// Incr off by length.
			off += int(ent.Reclen)

			// Skip the '.' and '..' dir entries.
			if ent.IsCurDir() || ent.IsPrevDir() {
				continue iter
			}

			// Pass entry to given function.
			if err := each(ent); err != nil {
				return err
			}
		}
	}

	return nil
}

// readdirent is a simple wrapper around syscall.ReadDirent().
func readdirent(fd int, buf []byte) (n int, err error) {
	err = retry_on_eintr(func() error {
		n, err = syscall.ReadDirent(fd, buf)
		return err
	})
	return
}

// open calls openfd() and wraps fd in os.File{}, adding
// the file descriptor to Go runtime's netpoll system.
func open(path string, args OpenArgs) (*os.File, error) {
	fd, err := open_fd(path, args)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

// openfd is a simple wrapper around syscall.Open().
func open_fd(path string, args OpenArgs) (fd int, err error) {
	err = retry_on_eintr(func() (err error) {
		fd, err = syscall.Open(path, args.Flags, args.Perms)
		return
	})
	return
}

// stat is a simple wrapper around syscall.Stat().
func stat(path string, stat *syscall.Stat_t) error {
	return retry_on_eintr(func() error {
		return syscall.Stat(path, stat)
	})
}

// lstat is a simple wrapper around syscall.Lstat().
func lstat(path string, stat *syscall.Stat_t) error {
	return retry_on_eintr(func() error {
		return syscall.Lstat(path, stat)
	})
}

// chown is a simple wrapper around syscall.Chown().
func chown(path string, uid, gid int) error {
	return retry_on_eintr(func() error {
		return syscall.Chown(path, uid, gid)
	})
}

// chmod is a simple wrapper around syscall.Chmod().
func chmod(path string, mode uint32) error {
	return retry_on_eintr(func() error {
		return syscall.Chmod(path, mode)
	})
}

// rename is a simple wrapper around syscall.Rename().
func rename(oldpath, newpath string) error {
	return retry_on_eintr(func() error {
		return syscall.Rename(oldpath, newpath)
	})
}

// unlink is a simple wrapper around syscall.Unlink().
func unlink(path string) error {
	return retry_on_eintr(func() error {
		return syscall.Unlink(path)
	})
}

// rmdir is a simple wrapper around syscall.Rmdir().
func rmdir(path string) error {
	return retry_on_eintr(func() error {
		return syscall.Rmdir(path)
	})
}

// symlink is a simple wrapper around syscall.Symlink().
func symlink(oldpath, newpath string) error {
	return retry_on_eintr(func() error {
		return syscall.Symlink(oldpath, newpath)
	})
}

// link is a simple wrapper around syscall.Link()
func link(oldpath, newpath string) error {
	return retry_on_eintr(func() error {
		return syscall.Link(oldpath, newpath)
	})
}

// close is a simple wrapper around syscall.Close().
func close(fd int) error {
	return retry_on_eintr(func() error {
		return syscall.Close(fd)
	})
}

// retry_on_eintr is a low-level filesystem function
// for retrying syscalls on O_EINTR received, i.e.
// the system call was interrupted and not resumed.
func retry_on_eintr(do func() error) error {
	for {
		err := do()
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}

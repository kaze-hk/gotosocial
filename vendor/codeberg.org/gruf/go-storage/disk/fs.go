package disk

import (
	"os"
	"strings"
	"syscall"

	"codeberg.org/gruf/go-fastpath/v2"
	"codeberg.org/gruf/go-storage"
	"codeberg.org/gruf/go-storage/internal"
)

// FS is a simple wrapper around a base directory
// path to provide file system operations within
// that directory. It also translates ENOENT and
// EEXIST errors to their equivalent storage errors.
//
// The uninitialized FS is safe to use,
// it will simply use the current dir.
type FS struct{ base string }

// NewFS returns a new FS{} with base path.
func NewFS(base string) FS { return FS{base} }

// String returns the defined FS{} base path.
func (fs FS) String() string { return fs.base }

// Open performs syscall.Open() on the file at relative path, with given OpenArgs{}.
//
// NOTE: this does not perform much of the wrapping that os.OpenFile() does, it may
// not set appropriate arguments for opening files other than regular / directories!
func (fs FS) Open(path string, args OpenArgs) (*os.File, error) {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return nil, err
	}

	// Open file path with args.
	file, err := open(path, args)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)

	case syscall.EEXIST:
		if args.Flags&syscall.O_EXCL != 0 {
			// Translate already exists errors and wrap with the path.
			err = internal.ErrWithKey(storage.ErrAlreadyExists, path)
		}
	}

	return file, err
}

// Chown performs syscall.Chown() on the file in FS at relative path.
func (fs FS) Chown(path string, uid, gid int) error {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return err
	}

	// Perform chmown on file.
	err = chown(path, uid, gid)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return err
}

// Chmod performs syscall.Chmod() on the file in FS at relative path.
func (fs FS) Chmod(path string, mode uint32) error {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return err
	}

	// Perform chmod on file.
	err = chmod(path, mode)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return err
}

// ReadDir gathers entries from WalkDir() and allocates a DirEntry{} for each.
func (fs FS) ReadDir(path string) ([]DirEntry, error) {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return nil, err
	}

	var entries []DirEntry

	// Gather entries as returnable DirEntry{} type.
	if err := readdir(path, func(ent *Dirent) error {
		entries = append(entries, DirEntry{
			Path: ent.NameStr(),
			Type: ent.Type,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return entries, nil
}

// Walk performs syscall.ReadDirent() on dir tree in FS at relative path, passing each entry
// to given function. NOTE: DIRENT MEMORY IS NOT SAFE FOR REUSE OUTSIDE OF EACH FUNCTION CALL.
func (fs FS) Walk(path string, each func(dir string, ent *Dirent) error) error {
	if each == nil {
		panic("nil func")
	}

	// Acquire path builder buffer.
	pb := internal.GetPathBuilder()
	defer internal.PutPathBuilder(pb)

	// Generate path from relative.
	path, err := fs.filepath(pb, path)
	if err != nil {
		return err
	}

	// Walk entire directory tree, only passing through the relative dir path.
	return walk_dir(pb, path, func(absdir, reldir string, ent *Dirent) error {
		return each(reldir, ent)
	})
}

// WalkDir performs syscall.ReadDirent() on dir in FS at relative path, passing each entry
// to given function. NOTE: DIRENT MEMORY IS NOT SAFE FOR REUSE OUTSIDE OF EACH FUNCTION CALL.
func (fs FS) WalkDir(path string, each func(ent *Dirent) error) error {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return err
	}

	// Read directory entries.
	err = readdir(path, each)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return err
}

// Stat performs syscall.Stat() on the file in FS at relative path.
func (fs FS) Stat(path string) (syscall.Stat_t, error) {
	var stat_t syscall.Stat_t

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return stat_t, err
	}

	// Stat file info on disk.
	err = stat(path, &stat_t)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return stat_t, err
}

// Lstat performs syscall.Lstat() on the file in FS at relative path.
func (fs FS) Lstat(path string) (syscall.Stat_t, error) {
	var stat_t syscall.Stat_t

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return stat_t, err
	}

	// Stat file info on disk.
	err = lstat(path, &stat_t)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return stat_t, err
}

// Unlink performs syscall.Unlink() on the file in FS at relative path.
func (fs FS) Unlink(path string) error {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return err
	}

	// Remove reg file.
	err = unlink(path)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return err
}

// Rmdir performs syscall.Rmdir() on the dir in FS at relative path.
func (fs FS) Rmdir(path string) error {

	// Generate path from relative.
	path, err := fs.Filepath(path)
	if err != nil {
		return err
	}

	// Remove dir file.
	err = rmdir(path)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, path)
	}

	return err
}

// Rename performs syscall.Rename() on the old and new paths in FS.
func (fs FS) Rename(oldpath, newpath string) error {

	// Acquire path builder buffer.
	pb := internal.GetPathBuilder()

	// Generate file path for old path.
	old, err1 := fs.filepath(pb, oldpath)

	// Generate file path for new path.
	new, err2 := fs.filepath(pb, newpath)

	// Done with path buffer.
	internal.PutPathBuilder(pb)

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	// Rename old to new file.
	err := rename(old, new)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, oldpath)
	}

	return err
}

// Symlink performs syscall.Symlink() on the source and destination paths in FS.
func (fs FS) Symlink(oldpath, newpath string) error {

	// Acquire path builder buffer.
	pb := internal.GetPathBuilder()

	// Generate file path for old path.
	old, err1 := fs.filepath(pb, oldpath)

	// Generate file path for new path.
	new, err2 := fs.filepath(pb, newpath)

	// Done with path buffer.
	internal.PutPathBuilder(pb)

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	// Create disk symlink.
	err := symlink(old, new)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, oldpath)
	}

	return err
}

// Link performs syscall.Link() on the source and destination paths in FS.
func (fs FS) Link(oldpath, newpath string) error {

	// Acquire path builder buffer.
	pb := internal.GetPathBuilder()

	// Generate file path for old path.
	old, err1 := fs.filepath(pb, oldpath)

	// Generate file path for new path.
	new, err2 := fs.filepath(pb, newpath)

	// Done with path buffer.
	internal.PutPathBuilder(pb)

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	// Create disk hardlink.
	err := link(old, new)
	switch err {

	case syscall.ENOENT:
		// Translate not-found errors and wrap with the path.
		err = internal.ErrWithKey(storage.ErrNotFound, oldpath)
	}

	return err
}

// Filepath checks and returns a cleaned filepath within FS{} base.
func (fs FS) Filepath(path string) (string, error) {
	pb := internal.GetPathBuilder()
	path, err := fs.filepath(pb, path)
	internal.PutPathBuilder(pb)
	return path, err
}

// filepath performs the "meat" of Filepath(), working with an existing fastpath.Builder{}.
func (fs FS) filepath(pb *fastpath.Builder, path string) (string, error) {
	old := path

	// Build from base.
	pb.Append(fs.base)
	pb.Append(path)

	// Take COPY of bytes.
	path = string(pb.B)

	// Check for dir traversal outside base.
	if isDirTraversal(fs.base, path) {
		return "", internal.ErrWithKey(storage.ErrInvalidKey, old)
	}

	return path, nil
}

// isDirTraversal will check if rootPlusPath is a dir traversal outside of root,
// assuming that both are cleaned and that rootPlusPath is path.Join(root, somePath).
func isDirTraversal(root, rootPlusPath string) bool {
	switch root {

	// Root is $PWD, check
	// for traversal out of
	case "", ".":
		return strings.HasPrefix(rootPlusPath, "../")

	// Root is *root*, ensure
	// it's not trying escape
	case "/":
		switch l := len(rootPlusPath); {
		case l == 3: // i.e. root=/ plusPath=/..
			return rootPlusPath[:3] == "/.."
		case l >= 4: // i.e. root=/ plusPath=/../[etc]
			return rootPlusPath[:4] == "/../"
		}
		return false
	}
	switch {

	// The path MUST be prefixed by storage root
	case !strings.HasPrefix(rootPlusPath, root):
		return true

	// In all other cases,
	// check not equal
	default:
		return len(root) == len(rootPlusPath)
	}
}

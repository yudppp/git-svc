package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/git-svc/pkg/utils" // Adjust module path
	"github.com/spf13/cobra"
	// Consider using a table printer library if output becomes complex, e.g., "github.com/olekukonko/tablewriter"
	// For now, fmt.Printf is fine for MVP.
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed service worktrees and symlinks",
	Long:  `Displays a list of all symbolic links managed by git-svc and their corresponding branches/worktrees.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := utils.GetGitRepoRoot()
		if err != nil {
			return fmt.Errorf("failed to get git repo root: %w", err)
		}

		// CurrentWorktreeRoot is now set by RootCmd's PersistentPreRunE
		absWorktreeRootPath := filepath.Join(repoRoot, CurrentWorktreeRoot)
		// Normalize absWorktreeRootPath to clean it up (e.g. remove trailing slashes) for reliable string comparisons later
		absWorktreeRootPath = filepath.Clean(absWorktreeRootPath)


		var managedLinks [][2]string // Store pairs of [symlinkDisplayPath, branch]

		err = filepath.WalkDir(repoRoot, func(currentPath string, d fs.DirEntry, errWalk error) error {
			if errWalk != nil {
				// If there's an error accessing a path, print it and continue if possible
				fmt.Fprintf(os.Stderr, "Warning: error accessing %s: %v\n", currentPath, errWalk)
				// Decide if to skip or stop. For robust listing, try to continue.
				if d != nil && d.IsDir() {
					// Potentially skip the directory if error is like permission denied
					// return filepath.SkipDir
				}
				return nil // Continue walking if possible
			}

			// Ensure currentPath is absolute for reliable processing
			if !filepath.IsAbs(currentPath) {
				// This case should ideally not happen if repoRoot is absolute.
				// WalkDir usually provides absolute paths if the root is absolute.
				// However, being defensive here.
				currentPath = filepath.Join(repoRoot, currentPath)
			}
			currentPath = filepath.Clean(currentPath)


			// Skip .git directory
			if d.IsDir() && d.Name() == ".git" {
				return filepath.SkipDir
			}

			// Skip the worktree root directory itself and anything underneath it.
			// We are looking for symlinks *outside* WorktreeRoot pointing *into* it.
			if strings.HasPrefix(currentPath, absWorktreeRootPath) {
				// If currentPath is absWorktreeRootPath itself, skip its contents.
				// If currentPath is deeper inside absWorktreeRootPath, it's not a candidate symlink location.
				if currentPath == absWorktreeRootPath && d.IsDir() {
					return filepath.SkipDir
				} else if currentPath != absWorktreeRootPath {
                    // If it's a file or dir strictly inside absWorktreeRootPath, skip.
                    // This prevents finding symlinks inside .worktrees that might point elsewhere.
                    return nil
                }
			}


			if d.Type()&os.ModeSymlink == os.ModeSymlink {
				symlinkAbsPath := currentPath // currentPath is already absolute and cleaned

				symlinkDisplayPath, errRel := filepath.Rel(repoRoot, symlinkAbsPath)
				if errRel != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not get relative path for symlink %s: %v\n", symlinkAbsPath, errRel)
					return nil // Skip this symlink
				}

				targetAbsPath, errResolve := utils.ResolveSymlink(symlinkAbsPath)
				if errResolve != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve symlink %s (points to %s): %v\n", symlinkDisplayPath, targetAbsPath, errResolve)
					return nil
				}
                targetAbsPath = filepath.Clean(targetAbsPath)


				// Check if targetAbsPath is within absWorktreeRootPath
				if strings.HasPrefix(targetAbsPath, absWorktreeRootPath+string(filepath.Separator)) {
					pathRelativeToWorktreeRoot, errRelTarget := filepath.Rel(absWorktreeRootPath, targetAbsPath)
					if errRelTarget == nil && !strings.HasPrefix(pathRelativeToWorktreeRoot, "..") {

						parts := strings.SplitN(pathRelativeToWorktreeRoot, string(filepath.Separator), 2)

						if len(parts) >= 1 {
							branchName := parts[0]
							// Prevent empty branch names or names that are just separators (unlikely with filepath.Clean)
							if branchName == "" || branchName == "." || strings.Contains(branchName, string(filepath.Separator)) {
                                // fmt.Fprintf(os.Stderr, "Debug: Invalid branch name '%s' for symlink %s\n", branchName, symlinkDisplayPath)
                                return nil // Invalid branch name extracted
                            }

							expectedTargetInnerStructure := ""
							if len(parts) == 2 {
								expectedTargetInnerStructure = parts[1]
							}

							normalizedSymlinkDisplayPath := filepath.ToSlash(symlinkDisplayPath)
							normalizedExpectedTargetInnerStructure := filepath.ToSlash(expectedTargetInnerStructure)

							if normalizedExpectedTargetInnerStructure == normalizedSymlinkDisplayPath {
								managedLinks = append(managedLinks, [2]string{symlinkDisplayPath, branchName})
							} else {
								// Optional: For debugging
								// fmt.Fprintf(os.Stderr, "Debug: Symlink %s (target %s) points into worktree %s, but structure mismatch. Expected structure: %s, actual structure in worktree: %s\n",
								// symlinkDisplayPath, targetAbsPath, branchName, normalizedSymlinkDisplayPath, normalizedExpectedTargetInnerStructure)
							}
						}
					}
				}
			}
			return nil
		})

		if err != nil {
			// This error is from WalkDir if it encounters an issue it cannot recover from based on the callback's returns.
			return fmt.Errorf("error walking directory %s: %w", repoRoot, err)
		}

		if len(managedLinks) == 0 {
			fmt.Println("No git-svc managed symlinks found.")
			return nil
		}

		fmt.Println("Managed symlinks:")
		fmt.Printf("%-40s %-20s\n", "SYMLINK (relative to repo root)", "BRANCH")
		fmt.Println(strings.Repeat("-", 61)) // Adjusted for slight title change
		for _, link := range managedLinks {
			fmt.Printf("%-40s %-20s\n", link[0], link[1])
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

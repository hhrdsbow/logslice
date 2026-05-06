package slicer

import (
	"context"
)

// CheckpointStage wraps a line channel and periodically saves a checkpoint
// every saveEvery lines, allowing resumable processing.
type CheckpointStage struct {
	cp        *Checkpoint
	savePath  string
	saveEvery int64
}

// NewCheckpointStage creates a stage that saves a checkpoint every saveEvery
// lines to savePath. Use saveEvery <= 0 to disable periodic saving.
func NewCheckpointStage(cp *Checkpoint, savePath string, saveEvery int64) *CheckpointStage {
	if saveEvery <= 0 {
		saveEvery = 1000
	}
	return &CheckpointStage{
		cp:        cp,
		savePath:  savePath,
		saveEvery: saveEvery,
	}
}

// Run reads lines from in, updates the checkpoint, saves periodically, and
// forwards each line to the returned channel. It closes the output channel
// when in is closed or ctx is done.
func (s *CheckpointStage) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string, cap(in))
	go func() {
		defer close(out)
		var lineCount int64
		var byteOffset int64
		for {
			select {
			case <-ctx.Done():
				_ = s.cp.Save(s.savePath)
				return
			case line, ok := <-in:
				if !ok {
					_ = s.cp.Save(s.savePath)
					return
				}
				lineCount++
				byteOffset += int64(len(line)) + 1 // +1 for newline
				s.cp.Update(byteOffset, lineCount)
				if lineCount%s.saveEvery == 0 {
					_ = s.cp.Save(s.savePath)
				}
				select {
				case out <- line:
				case <-ctx.Done():
					_ = s.cp.Save(s.savePath)
					return
				}
			}
		}
	}()
	return out
}

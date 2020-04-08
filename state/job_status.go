package state

import (
	"github.com/hashicorp/nomad-autoscaler/state/status"
	"github.com/hashicorp/nomad/api"
)

// jobStatusUpdate is responsible for taking a Nomad JobScaleStatusResponse and
// updating the autoscaler internal state.
func (h *Handler) jobStatusUpdate(update *api.JobScaleStatusResponse) {

	// If the job is stopped, remove the scaling policies, remove the scale
	// status handler and delete the scale status state. Even if the job is
	// started again, the Autoscaler will gather all the information it
	// needs so we can clear everything out.
	if update.JobStopped {
		h.PolicyState.DeletePolicies(update.JobID)
		h.removeJobStatusHandler(update.JobID)
		h.statusState.DeleteJob(update.JobID)
		h.log.Debug("job is stopped, removed internal state", "job_id", update.JobID)
	} else {
		// Store the new scale status information and move along.
		h.statusState.SetJob(update)
		h.log.Debug("set scale status in state", "job_id", update.JobID)
	}
}

// startJobStatusWatcher will start a new job scale status blocking watcher if
// needed. It will also wait for the first run of the watcher to process the
// API data, which is required before processing a scaling policy.
func (h *Handler) startJobStatusWatcher(jobID string) {
	h.statusWatcherHandlerLock.Lock()
	defer h.statusWatcherHandlerLock.Unlock()

	if _, ok := h.statusWatcherHandlers[jobID]; ok {
		h.log.Trace("found existing handler for job scale status", "job_id", jobID)
		return
	}
	h.log.Debug("starting new handler for job scale status", "job_id", jobID)

	// Setup the new handler, start it and wait for its initial run to finish.
	h.statusWatcherHandlers[jobID] = status.NewWatcher(h.log, jobID, h.nomad, h.jobStatusUpdate)
	go h.statusWatcherHandlers[jobID].Start()
	<-h.statusWatcherHandlers[jobID].InitialLoad
}

// removeJobStatusHandler will remove the scale status handler for the job from
// the state handlers mapping.
//
// TODO(jrasell) this should shutdown the blocking query; currently the query
//  will just reach the timeout.
func (h *Handler) removeJobStatusHandler(jobID string) {
	h.statusWatcherHandlerLock.Lock()
	defer h.statusWatcherHandlerLock.Unlock()

	// Delete the handle.
	delete(h.statusWatcherHandlers, jobID)
	h.log.Debug("deleted scale status watcher handle", "job_id", jobID)
}
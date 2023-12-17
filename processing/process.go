// Summary: Process a job
// Description: This file is used as a entry for job processing. It fetches an unclaimed job, validates it, and processes it using a corresponding function.
// Author: Egor Pristavka
// Date: 2023-03-27

package processing

import (
	"context"
	sm "dev.hackerman.me/artheon/veverse-shared/model"
	"fmt"
	"l7-cloud-builder/api"
	"l7-cloud-builder/config"
	"l7-cloud-builder/logger"
	"time"
)

func Process(ctx context.Context) (err error) {

	//region Wait for an unclaimed job

	var job *sm.JobV2
	job, err = api.FetchUnclaimedJob(ctx)
	if err != nil {
		return err
	}

	// Wait for jobs to be scheduled
	if job == nil {
		// Wait before the next request
		time.Sleep(10 * time.Second)
		return nil
	}

	//endregion

	//region Defer job status update

	defer func(job *sm.JobV2) {
		if job != nil {
			if err != nil {
				if err1 := api.UpdateJobStatus(ctx, job, config.JobStatusError, err.Error()); err1 != nil {
					logger.Logger.Errorf("failed to update job status: %v", err1)
				}
			} else {
				if err1 := api.UpdateJobStatus(ctx, job, config.JobStatusCompleted, ""); err1 != nil {
					logger.Logger.Errorf("failed to update job status: %v", err1)
				}
			}
		} else {
			logger.Logger.Errorf("failed to update job status, job is nil")
		}
	}(job)

	//endregion

	//region Validate the received job

	// Validate job type
	if !config.Config.EnabledJobs[job.Type] {
		err = fmt.Errorf("invalid job type: %s", job.Type)
		return
	}

	// Validate job deployment
	if !config.Config.EnabledTargets[job.Target] {
		err = fmt.Errorf("invalid job deployment: %s", job.Target)
		return
	}

	// Validate job platform
	if !config.Config.EnabledPlatforms[job.Platform] {
		err = fmt.Errorf("invalid job platform: %s", job.Platform)
		return
	}

	//endregion

	//region Process the job

	if job.Type == config.Config.JobMapping[config.JobTypeRelease] {
		// Requires the release to be set for the job
		if job.Release == nil {
			return fmt.Errorf("no release metadata, required for job type: %s", job.Type)
		}

		if job.Target == config.Config.TargetMapping[config.TargetTypeClient] {
			err = processReleaseClient(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypeServer] {
			err = processReleaseServer(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypeEditor] {
			err = processReleaseEditor(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypeLauncher] {
			err = processReleaseLauncher(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypeServerLauncher] {
			err = processReleaseServerLauncher(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypePixelStreamingLauncher] {
			err = processReleasePixelStreamingLauncher(*job)
		} else {
			err = fmt.Errorf("invalid job deployment %s for type %s", job.Target, job.Type)
		}
	} else if job.Type == config.Config.JobMapping[config.JobTypePackage] {
		// Requires the package to be set for the job
		if job.Package == nil {
			return fmt.Errorf("no package metadata, required for job type: %s", job.Type)
		}

		if job.Target == config.Config.TargetMapping[config.TargetTypeClient] {
			err = processPackageClient(*job)
		} else if job.Target == config.Config.TargetMapping[config.TargetTypeServer] {
			err = processPackageServer(*job)
		} else {
			err = fmt.Errorf("invalid job deployment %s for type %s", job.Target, job.Type)
		}
	} else {
		err = fmt.Errorf("invalid job type: %s", job.Type)
	}

	//endregion

	return err
}

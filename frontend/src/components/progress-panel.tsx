import { InstallState, Mode } from '../hooks/use-installer'

interface ProgressPanelProps {
	mode: Mode
	progress: number
	progressLabel: string
	installState: InstallState
}

export function ProgressPanel(props: ProgressPanelProps) {
	const { mode, progress, progressLabel } = props

	const isInstallingApp = mode === 'install'

	return (
		<div className='wrapper-block-progress'>
			{isInstallingApp && (
				<div className='progress-block'>
					<div className='progress-header'>
						<span>install progress</span>
						<span className='progress-percent'>{`${progress}%`}</span>
					</div>
					<div className='progress-bar'>
						<div
							className='progress-fill'
							style={{
								width: `${progress}%`,
							}}
						/>
					</div>
					<div className='progress-label'>{progressLabel}</div>
				</div>
			)}
		</div>
	)
}

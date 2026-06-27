import { useEffect, useRef, useState } from 'react'
import { RunBundledInstaller } from '../../wailsjs/go/installer/Installer'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export type InstallState = 'idle' | 'running' | 'done'
export type Mode = 'install' | 'remove'

export function useInstaller() {
	const [installState, setInstallState] = useState<InstallState>('idle')
	const [mode] = useState<Mode>('install')
	const [progress, setProgress] = useState(0)
	const [statusLabel, setStatusLabel] = useState('ready')
	const [progressLabel, setProgressLabel] = useState('Idle')

	const timerRef = useRef<number | null>(null)

	useEffect(() => {
		const offText = EventsOn('install:progressText', (data: any) => {
			if (typeof data === 'string' && data.trim().length > 0) {
				setProgressLabel(data)
			}
		})

		const start = async () => {
			setInstallState('running')
			setStatusLabel('installing')
			setProgress(0)
			setProgressLabel('Installing…')

			if (timerRef.current !== null) {
				clearInterval(timerRef.current)
				timerRef.current = null
			}

			timerRef.current = window.setInterval(() => {
				setProgress(prev => {
					if (prev >= 100) return 100
					return prev + 2
				})
			}, 120)

			try {
				await RunBundledInstaller()
			} catch {}
		}

		start()

		return () => {
			offText()
			if (timerRef.current !== null) {
				clearInterval(timerRef.current)
				timerRef.current = null
			}
		}
	}, [])

	useEffect(() => {
		if (progress >= 100 && installState === 'running') {
			setStatusLabel('waiting')
			setProgressLabel('Waiting')
			setInstallState('done')
			if (timerRef.current !== null) {
				clearInterval(timerRef.current)
				timerRef.current = null
			}
		}
	}, [progress, installState])

	const statusDotClass =
		statusLabel === 'completed'
			? 'status-dot dot-success'
			: installState === 'running'
				? 'status-dot dot-warn'
				: 'status-dot'

	return {
		installState,
		mode,
		progress,
		statusLabel,
		progressLabel,
		isFetching: false,
		downloadPath: null,
		removeStatus: 'Disabled',
		updateInfo: null,
		isCheckingUpdate: false,
		isInstallingUpdate: false,
		isUpdateModalOpen: false,
		statusDotClass,
		nextButtonLabel: '',
		hasUpdate: false,
		overlayOpenAttr: 'false',
		handleNextClick: () => {},
		handleRefreshClick: () => {},
		runUpdateInstallFlow: async () => {},
		setIsUpdateModalOpen: () => {},
		removePath: 'Disabled',
	}
}

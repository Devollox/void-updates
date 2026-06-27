import './App.css'
import { Header } from './components/header'
import { ProgressPanel } from './components/progress-panel'
import { useInstaller } from './hooks/use-installer'

function App() {
	const state = useInstaller()

	return (
		<div className='AppRoot'>
			<div className='installer-shell'>
				<Header statusLabel={state.statusLabel} statusDotClass={state.statusDotClass} />

				<div className='installer-main'>
					<div className='left-pane'>
						<ProgressPanel
							mode={state.mode}
							progress={state.progress}
							progressLabel={state.progressLabel}
							installState={state.installState}
						/>
					</div>
				</div>
			</div>
		</div>
	)
}

export default App

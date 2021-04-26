import logo from './logo.svg';
import './App.css';
import 'primeflex/primeflex.css';
import 'primereact/resources/themes/saga-blue/theme.css';
import 'primereact/resources/primereact.min.css';
import 'primeicons/primeicons.css';
import {Dashboard} from "./pages/dashboard";
import {HashRouter} from "react-router-dom";

function App() {
  return (
    <HashRouter>
      <div className="App">
        <Dashboard/>
      </div>
    </HashRouter>
  );
}

export default App;

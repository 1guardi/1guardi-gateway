import { BrowserRouter, Routes, Route } from 'react-router-dom'
import CompanyIntro from './CompanyIntro'
import AIGatewayLanding from './AIGatewayLanding'
import './index.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<CompanyIntro />} />
        <Route path="/gateway" element={<AIGatewayLanding />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App

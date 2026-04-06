import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Host from "./routes/Host";
import Play from "./routes/Play";
import Display from "./routes/Display";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Host />} />
        <Route path="/play" element={<Play />} />
        <Route path="/display" element={<Display />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);

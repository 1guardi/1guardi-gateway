import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from registry import load_registry

app = FastAPI(title="ML Runner", version="1.0.0")

config_path = os.getenv("ANALYZERS_CONFIG", "analyzers.yml")
REGISTRY = load_registry(config_path)


class AnalyzeRequest(BaseModel):
    text: str


@app.get("/health")
def health():
    return {"status": "ok", "analyzers": list(REGISTRY.keys())}


@app.post("/analyze/{analyzer_name}")
def analyze(analyzer_name: str, req: AnalyzeRequest):
    if analyzer_name not in REGISTRY:
        raise HTTPException(
            status_code=404,
            detail=f"unknown analyzer: {analyzer_name!r}. available: {list(REGISTRY.keys())}",
        )

    entry = REGISTRY[analyzer_name]
    pipe = entry["pipeline"]
    max_len = entry["max_length"]

    result = pipe(req.text[:max_len], truncation=True)
    return {"analyzer": analyzer_name, "result": result}

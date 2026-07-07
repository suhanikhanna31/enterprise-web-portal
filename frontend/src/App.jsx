import React, { useState, useEffect } from 'react';

const API_BASE = 'http://localhost:8080/api';

function App() {
  const [campaigns, setCampaigns] = useState([]);
  const [analytics, setAnalytics] = useState({ impressions: 0, clicks: 0, ctr: 0, revenue: 0 });
  const [name, setName] = useState('');
  const [budget, setBudget] = useState('');
  const [loadingTraffic, setLoadingTraffic] = useState(false);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setError(null);
    try {
      const campRes = await fetch(`${API_BASE}/campaigns`);
      if (!campRes.ok) throw new Error(`Failed to load campaigns (${campRes.status})`);
      setCampaigns(await campRes.json());

      const analyticRes = await fetch(`${API_BASE}/analytics`);
      if (!analyticRes.ok) throw new Error(`Failed to load analytics (${analyticRes.status})`);
      setAnalytics(await analyticRes.json());
    } catch (err) {
      console.error('Error communicating with backend:', err);
      setError('Could not reach the backend. Is the Go server running on :8080?');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateCampaign = async (e) => {
    e.preventDefault();
    if (!name || !budget) return;

    try {
      const res = await fetch(`${API_BASE}/campaigns`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, budget: parseFloat(budget) }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `Request failed (${res.status})`);
      }
      setName('');
      setBudget('');
      fetchData();
    } catch (err) {
      setError(err.message);
    }
  };

  // Generates test events against a campaign so the dashboard has data
  // to show. This is sample-data generation for local development, not
  // a load-testing or traffic-simulation tool.
  const generateSampleEvents = async (campaignId) => {
    setLoadingTraffic(true);
    setError(null);
    try {
      for (let i = 0; i < 50; i++) {
        const type = Math.random() > 0.85 ? 'click' : 'impression';
        const res = await fetch(`${API_BASE}/track?campaign_id=${campaignId}&type=${type}`, { method: 'POST' });
        if (res.status === 402) {
          // Campaign budget exhausted mid-loop; stop early rather than
          // hammering a paused campaign.
          break;
        }
      }
      await fetchData();
    } catch (err) {
      setError('Failed to generate sample events.');
    } finally {
      setLoadingTraffic(false);
    }
  };

  return (
    <div className="min-h-screen p-8">
      <header className="border-b border-slate-800 pb-6 mb-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white">Ad Campaign Dashboard</h1>
          <p className="text-sm text-slate-400 mt-1">A small demo app: create campaigns, log events, view basic stats.</p>
        </div>
        <div className="bg-slate-500/10 border border-slate-500/20 px-3 py-1 rounded text-slate-400 text-xs font-mono">
          Demo build
        </div>
      </header>

      {error && (
        <div className="mb-6 bg-red-500/10 border border-red-500/30 text-red-300 text-sm px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {isLoading ? (
        <p className="text-slate-500 text-sm">Loading...</p>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Impressions</span>
              <div className="text-3xl font-bold mt-2 text-indigo-400">{analytics.impressions.toLocaleString()}</div>
            </div>
            <div className="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Clicks</span>
              <div className="text-3xl font-bold mt-2 text-sky-400">{analytics.clicks.toLocaleString()}</div>
            </div>
            <div className="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">CTR</span>
              <div className="text-3xl font-bold mt-2 text-amber-400">{analytics.ctr.toFixed(2)}%</div>
            </div>
            <div className="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Revenue (simulated)</span>
              <div className="text-3xl font-bold mt-2 text-emerald-400">${analytics.revenue.toFixed(2)}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <div className="bg-slate-800/30 border border-slate-800 p-6 rounded-xl h-fit">
              <h2 className="text-xl font-bold mb-4 text-white">New Campaign</h2>
              <form onSubmit={handleCreateCampaign} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Name</label>
                  <input type="text" value={name} onChange={e => setName(e.target.value)} className="w-full bg-slate-900 border border-slate-700 rounded px-3 py-2 text-white focus:outline-none focus:border-indigo-500" placeholder="Q3 Brand Awareness" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Budget ($)</label>
                  <input type="number" value={budget} onChange={e => setBudget(e.target.value)} className="w-full bg-slate-900 border border-slate-700 rounded px-3 py-2 text-white focus:outline-none focus:border-indigo-500" placeholder="50000" />
                </div>
                <button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white py-2 rounded font-semibold transition-colors">
                  Create Campaign
                </button>
              </form>
            </div>

            <div className="lg:col-span-2 bg-slate-800/30 border border-slate-800 p-6 rounded-xl">
              <h2 className="text-xl font-bold mb-4 text-white">Campaigns</h2>
              {campaigns.length === 0 ? (
                <p className="text-slate-500 text-sm italic">No campaigns yet.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm text-slate-300">
                    <thead className="bg-slate-900/50 text-xs uppercase text-slate-400 border-b border-slate-800">
                      <tr>
                        <th className="py-3 px-4">ID</th>
                        <th className="py-3 px-4">Name</th>
                        <th className="py-3 px-4">Budget</th>
                        <th className="py-3 px-4">Spent</th>
                        <th className="py-3 px-4">Status</th>
                        <th className="py-3 px-4 text-right">Sample Data</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {campaigns.map((c) => (
                        <tr key={c.id} className="hover:bg-slate-800/20 transition-colors">
                          <td className="py-3 px-4 font-mono text-slate-500">#{c.id}</td>
                          <td className="py-3 px-4 font-semibold text-white">{c.name}</td>
                          <td className="py-3 px-4 font-mono text-slate-400">${parseFloat(c.budget).toLocaleString()}</td>
                          <td className="py-3 px-4 font-mono text-slate-400">${parseFloat(c.spent || 0).toFixed(2)}</td>
                          <td className="py-3 px-4">
                            <span className={`px-2 py-0.5 rounded text-xs font-mono ${c.status === 'active' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-600/20 text-slate-400'}`}>
                              {c.status}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <button disabled={loadingTraffic || c.status !== 'active'} onClick={() => generateSampleEvents(c.id)} className="text-xs bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-300 px-3 py-1 rounded disabled:opacity-40 transition-all">
                              {loadingTraffic ? 'Generating...' : 'Generate 50 events'}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default App;

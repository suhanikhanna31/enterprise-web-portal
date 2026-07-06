import React, { useState, useEffect } from 'react';

const API_BASE = 'http://localhost:8080/api';

function App() {
  const [campaigns, setCampaigns] = useState([]);
  const [analytics, setAnalytics] = useState({ impressions: 0, clicks: 0, ctr: 0, revenue: 0 });
  const [name, setName] = useState('');
  const [budget, setBudget] = useState('');
  const [loadingTraffic, setLoadingTraffic] = useState(false);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const campRes = await fetch(`${API_BASE}/campaigns`);
      const campData = await campRes.json();
      setCampaigns(campData);

      const analyticRes = await fetch(`${API_BASE}/analytics`);
      const analyticData = await analyticRes.json();
      setAnalytics(analyticData);
    } catch (err) {
      console.error("Error communicating with Go backend backend:", err);
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
      if (res.ok) {
        setName('');
        setBudget('');
        fetchData();
      }
    } catch (err) {
      console.error(err);
    }
  };

  // Simulated traffic tool to show "Scale and Ambiguity Mindset"
  const simulateTraffic = async (campaignId) => {
    setLoadingTraffic(true);
    // Simulate bulk streams of impressions and random clicks
    for (let i = 0; i < 50; i++) {
      const type = Math.random() > 0.85 ? 'click' : 'impression';
      await fetch(`${API_BASE}/track?campaign_id=${campaignId}&type=${type}`, { method: 'POST' });
    }
    await fetchData();
    setLoadingTraffic(false);
  };

  return (
    <div class="min-h-screen p-8">
      <header class="border-b border-slate-800 pb-6 mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-extrabold tracking-tight text-white">AdTech Core Engineering Portal</h1>
          <p class="text-sm text-slate-400 mt-1">High-throughput distribution matrix & live publisher metrics</p>
        </div>
        <div class="bg-emerald-500/10 border border-emerald-500/20 px-3 py-1 rounded text-emerald-400 text-xs font-mono">
          Go Production Pipeline Online
        </div>
      </header>

      {/* Analytics Top Cards */}
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
          <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total Aggregated Impressions</span>
          <div class="text-3xl font-bold mt-2 text-indigo-400">{analytics.impressions.toLocaleString()}</div>
        </div>
        <div class="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
          <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">System Click Events</span>
          <div class="text-3xl font-bold mt-2 text-sky-400">{analytics.clicks.toLocaleString()}</div>
        </div>
        <div class="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
          <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Average CTR</span>
          <div class="text-3xl font-bold mt-2 text-amber-400">{analytics.ctr.toFixed(2)}%</div>
        </div>
        <div class="bg-slate-800/50 border border-slate-700/50 p-6 rounded-xl">
          <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Live Revenue Ledger</span>
          <div class="text-3xl font-bold mt-2 text-emerald-400">${analytics.revenue.toFixed(2)}</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Creation Panel */}
        <div class="bg-slate-800/30 border border-slate-800 p-6 rounded-xl h-fit">
          <h2 class="text-xl font-bold mb-4 text-white">Deploy New Ad Space</h2>
          <form onSubmit={handleCreateCampaign} class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Campaign Namespace</label>
              <input type="text" value={name} onChange={e => setName(e.target.value)} class="w-full bg-slate-900 border border-slate-700 rounded px-3 py-2 text-white focus:outline-none focus:border-indigo-500" placeholder="Q3 TopFunnel Brand Awareness" />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Relational Safe Budget Limit ($)</label>
              <input type="number" value={budget} onChange={e => setBudget(e.target.value)} class="w-full bg-slate-900 border border-slate-700 rounded px-3 py-2 text-white focus:outline-none focus:border-indigo-500" placeholder="50000" />
            </div>
            <button type="submit" class="w-full bg-indigo-600 hover:bg-indigo-500 text-white py-2 rounded font-semibold transition-colors">
              Provision Ledger Infrastructure
            </button>
          </form>
        </div>

        {/* Live System Grid View */}
        <div class="lg:col-span-2 bg-slate-800/30 border border-slate-800 p-6 rounded-xl">
          <h2 class="text-xl font-bold mb-4 text-white">Active System Ad Inventories</h2>
          {campaigns.length === 0 ? (
            <p class="text-slate-500 text-sm italic">No campaigns allocated in database memory pool yet.</p>
          ) : (
            <div class="overflow-x-auto">
              <table class="w-full text-left text-sm text-slate-300">
                <thead class="bg-slate-900/50 text-xs uppercase text-slate-400 border-b border-slate-800">
                  <tr>
                    <th class="py-3 px-4">UID</th>
                    <th class="py-3 px-4">Identifier</th>
                    <th class="py-3 px-4">Threshold allocation</th>
                    <th class="py-3 px-4">State</th>
                    <th class="py-3 px-4 text-right">Synthetic Performance Stress</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-800/60">
                  {campaigns.map((c) => (
                    <tr key={c.id} class="hover:bg-slate-800/20 transition-colors">
                      <td class="py-3 px-4 font-mono text-slate-500">#{c.id}</td>
                      <td class="py-3 px-4 font-semibold text-white">{c.name}</td>
                      <td class="py-3 px-4 font-mono text-slate-400">${parseFloat(c.budget).toLocaleString()}</td>
                      <td class="py-3 px-4">
                        <span class="px-2 py-0.5 rounded text-xs bg-emerald-500/10 text-emerald-400 font-mono">{c.status}</span>
                      </td>
                      <td class="py-3 px-4 text-right">
                        <button disabled={loadingTraffic} onClick={() => simulateTraffic(c.id)} class="text-xs bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-300 px-3 py-1 rounded disabled:opacity-40 transition-all">
                          {loadingTraffic ? 'Streaming...' : 'Inject 50 Live Requests'}
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
    </div>
  );
}

export default App;

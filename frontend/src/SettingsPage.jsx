import { useState, useEffect } from 'react';
import { listSettings, upsertSetting, deleteSetting } from './api.js';

function SettingRow({ setting, onEdit, onDelete }) {
  let parsed;
  try {
    parsed = JSON.parse(setting.setting_value);
  } catch { parsed = null; }

  return (
    <div className="bg-white border border-slate-200 rounded-xl p-4 flex gap-4 items-start justify-between">
      <div className="flex-1 min-w-0">
        <div className="font-semibold text-sm break-all mb-1">{setting.setting_key}</div>
        <div className="text-sm text-slate-500">
          {parsed ? (
            <pre className="bg-slate-50 p-2 rounded-md overflow-x-auto text-xs max-h-24">{JSON.stringify(parsed, null, 2)}</pre>
          ) : (
            <span>{setting.setting_value}</span>
          )}
        </div>
      </div>
      <div className="flex gap-2 shrink-0">
        <button className="px-3 py-1.5 rounded-md text-sm font-medium border border-slate-200 text-slate-700 hover:bg-slate-50 transition cursor-pointer" onClick={() => onEdit(setting)}>
          Edit
        </button>
        <button className="px-3 py-1.5 rounded-md text-sm font-medium bg-red-600 text-white hover:bg-red-700 transition cursor-pointer" onClick={() => onDelete(setting)}>
          Delete
        </button>
      </div>
    </div>
  );
}

function SettingForm({ initial, onSave, onCancel }) {
  const [key, setKey] = useState(initial?.setting_key || '');
  const [value, setValue] = useState(initial?.setting_value || '');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!key.trim()) return;
    setSaving(true);
    setError(null);
    try {
      await upsertSetting(key.trim(), value);
      onSave();
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const handlePretty = () => {
    try {
      const parsed = typeof value === 'string' ? JSON.parse(value) : value;
      setValue(JSON.stringify(parsed, null, 2));
    } catch {}
  };

  return (
    <form className="bg-white border border-slate-200 rounded-xl p-5 mb-5" onSubmit={handleSubmit}>
      <h3 className="text-lg font-semibold mb-4">{initial ? 'Edit Setting' : 'New Setting'}</h3>
      {error && <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm mb-3">{error}</div>}
      <div className="mb-3">
        <label htmlFor="setting-key" className="block text-xs font-medium text-slate-500 mb-1">Key</label>
        <input
          id="setting-key"
          className="w-full px-3 py-2 rounded-lg border border-slate-200 text-sm bg-white text-slate-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 focus:outline-none transition disabled:bg-slate-50 disabled:text-slate-500"
          value={key}
          onChange={(e) => setKey(e.target.value)}
          placeholder="setting_key"
          disabled={!!initial}
          required
        />
      </div>
      <div className="mb-3">
        <div className="flex justify-between items-center mb-1">
          <label htmlFor="setting-value" className="block text-xs font-medium text-slate-500">Value (JSON)</label>
          <button type="button" className="px-2.5 py-1 rounded-md text-xs font-medium border border-slate-200 text-slate-600 hover:bg-slate-50 transition cursor-pointer" onClick={handlePretty}>
            Pretty Print
          </button>
        </div>
        <textarea
          id="setting-value"
          className="w-full px-3 py-2 rounded-lg border border-slate-200 text-sm bg-white text-slate-900 font-mono leading-relaxed focus:border-blue-500 focus:ring-2 focus:ring-blue-200 focus:outline-none transition resize-y min-h-[100px]"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          rows={6}
          placeholder='{"key": "value"}'
          required
        />
      </div>
      <div className="flex gap-2 justify-end">
        <button
          className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          type="submit"
          disabled={saving}
        >
          {saving ? 'Saving\u2026' : 'Save'}
        </button>
        <button className="px-4 py-2 rounded-lg text-sm font-medium border border-slate-200 text-slate-700 hover:bg-slate-50 transition cursor-pointer" type="button" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </form>
  );
}

export default function SettingsPage() {
  const [settings, setSettings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [editing, setEditing] = useState(null);
  const [showNew, setShowNew] = useState(false);

  const fetchSettings = async () => {
    setLoading(true);
    setError(null);
    try {
      setSettings(await listSettings());
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchSettings(); }, []);

  const handleEdit = (setting) => {
    setEditing(setting);
    setShowNew(false);
  };

  const handleDelete = async (setting) => {
    if (!window.confirm(`Delete setting "${setting.setting_key}"?`)) return;
    try {
      await deleteSetting(setting.setting_key);
      fetchSettings();
    } catch (e) {
      setError(e.message);
    }
  };

  const handleSave = () => {
    setEditing(null);
    setShowNew(false);
    fetchSettings();
  };

  return (
    <div>
      <div className="flex flex-wrap gap-3 items-center mb-5">
        <h1 className="text-2xl font-bold tracking-tight">Settings</h1>
        <button
          className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          onClick={() => { setShowNew(true); setEditing(null); }}
          disabled={showNew || editing}
        >
          + New Setting
        </button>
      </div>

      {error && <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm mb-4">{error}</div>}

      {editing && <SettingForm initial={editing} onSave={handleSave} onCancel={() => setEditing(null)} />}
      {showNew && <SettingForm onSave={handleSave} onCancel={() => setShowNew(false)} />}

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="size-10 border-2 border-slate-200 border-t-blue-600 rounded-full animate-spin" />
        </div>
      ) : settings.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <p>No settings found.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {settings.map((s) => (
            <SettingRow key={s.id} setting={s} onEdit={handleEdit} onDelete={handleDelete} />
          ))}
        </div>
      )}
    </div>
  );
}

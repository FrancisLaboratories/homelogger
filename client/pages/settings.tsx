"use client";

import React, {useEffect, useMemo, useState} from 'react';
import {Container, Button, Form, Row, Col, Alert, Spinner} from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import {SERVER_URL} from "@/lib/config";
import {useSettings} from '@/components/SettingsContext';
import {RegionalSettings} from '@/lib/formatters';

const SettingsPage: React.FC = () => {
    const {settings, loading, update, refresh} = useSettings();
    const [backupLoading, setBackupLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [options, setOptions] = useState<{
        locales: string[];
        languages: string[];
        currencies: string[];
        timeZones: string[];
        measurementSystems: string[];
        weekStartOptions: number[];
        dateFormats: string[];
        numberingSystems: string[];
    } | null>(null);

    const [formState, setFormState] = useState<RegionalSettings>(settings);

    useEffect(() => {
        setFormState(settings);
    }, [settings]);

    useEffect(() => {
        const loadOptions = async () => {
            try {
                const resp = await fetch(`${SERVER_URL}/settings/options`);
                if (!resp.ok) throw new Error('Failed to load options');
                const data = await resp.json();
                setOptions(data);
            } catch (err) {
                console.error('Error loading settings options:', err);
            }
        };
        void loadOptions();
    }, []);

    const handleDownloadBackup = async () => {
        setBackupLoading(true);
        try {
            const res = await fetch(`${SERVER_URL}/backup/download`);
            if (!res.ok) {
                throw new Error('Failed to download backup');
            }

            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            const timestamp = new Date().toISOString().slice(0,19).replaceAll(':','-');
            a.download = `homelogger-backup-${timestamp}.zip`;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error(err);
            alert('Error downloading backup. See console for details.');
        } finally {
            setBackupLoading(false);
        }
    };

    const handleSave = async () => {
        setError(null);
        setSuccess(null);
        setSaving(true);
        const updated = await update(formState);
        if (!updated) {
            setError('Failed to save settings.');
        } else {
            setSuccess('Settings saved.');
            await refresh();
        }
        setSaving(false);
    };

    const currencies = useMemo(() => options?.currencies ?? [settings.currency], [options, settings.currency]);
    const dateFormats = useMemo(
        () =>
            options?.dateFormats ?? [
                settings.dateFormat || 'YYYY-MM-DD',
            ],
        [options, settings.dateFormat]
    );
    const timeZones = useMemo(() => {
        if (typeof Intl !== 'undefined' && 'supportedValuesOf' in Intl) {
            try {
                // Use browser-provided list when available (full IANA set).
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const supported = (Intl as any).supportedValuesOf('timeZone') as string[];
                if (supported && supported.length > 0) return supported;
            } catch {
                // fall back to server list
            }
        }
        return options?.timeZones ?? [settings.timeZone];
    }, [options, settings.timeZone]);

    return (
        <Container>
            <MyNavbar />
            <h4 id='maintext' style={{marginTop: '1rem'}}>Settings</h4>
            {loading && (
                <div style={{marginTop: '1rem'}}>
                    <Spinner animation="border" size="sm" /> Loading settings...
                </div>
            )}
            {error && <Alert variant="danger" style={{marginTop: '1rem'}}>{error}</Alert>}
            {success && <Alert variant="success" style={{marginTop: '1rem'}}>{success}</Alert>}
            <div style={{marginTop: '1rem'}}>
                <h5>Regional Settings</h5>
                <Form>
                    <Row>
                        <Col md={6}>
                            <Form.Group controlId="formLocale">
                                <Form.Label>Locale</Form.Label>
                                <Form.Select
                                    value={formState.locale}
                                    onChange={(e) => setFormState({...formState, locale: e.target.value})}
                                >
                                    {(options?.locales ?? [settings.locale]).map((locale) => (
                                        <option key={locale} value={locale}>{locale}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                        <Col md={6}>
                            <Form.Group controlId="formLanguage">
                                <Form.Label>Language</Form.Label>
                                <Form.Select
                                    value={formState.language}
                                    onChange={(e) => setFormState({...formState, language: e.target.value})}
                                >
                                    {(options?.languages ?? [settings.language]).map((language) => (
                                        <option key={language} value={language}>{language}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                    </Row>
                    <Row>
                        <Col md={6}>
                            <Form.Group controlId="formCurrency">
                                <Form.Label>Currency</Form.Label>
                                <Form.Select
                                    value={formState.currency}
                                    onChange={(e) => setFormState({...formState, currency: e.target.value})}
                                >
                                    {currencies.map((currency) => (
                                        <option key={currency} value={currency}>{currency}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                        <Col md={6}>
                            <Form.Group controlId="formTimeZone">
                                <Form.Label>Time Zone</Form.Label>
                                <Form.Select
                                    value={formState.timeZone}
                                    onChange={(e) => setFormState({...formState, timeZone: e.target.value})}
                                >
                                    {timeZones.map((tz) => (
                                        <option key={tz} value={tz}>{tz}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                    </Row>
                    <Row>
                        <Col md={6}>
                            <Form.Group controlId="formMeasurementSystem">
                                <Form.Label>Measurement System</Form.Label>
                                <Form.Select
                                    value={formState.measurementSystem}
                                    onChange={(e) => setFormState({...formState, measurementSystem: e.target.value})}
                                >
                                    {(options?.measurementSystems ?? [settings.measurementSystem]).map((ms) => (
                                        <option key={ms} value={ms}>{ms}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                        <Col md={6}>
                            <Form.Group controlId="formWeekStart">
                                <Form.Label>Week Starts On</Form.Label>
                                <Form.Select
                                    value={formState.weekStart}
                                    onChange={(e) => setFormState({...formState, weekStart: Number(e.target.value)})}
                                >
                                    {(options?.weekStartOptions ?? [settings.weekStart]).map((value) => (
                                        <option key={value} value={value}>
                                            {value === 0 ? 'Sunday' : value === 1 ? 'Monday' : 'Saturday'}
                                        </option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                    </Row>
                    <Row>
                        <Col md={6}>
                            <Form.Group controlId="formDateFormat">
                                <Form.Label>Date Format</Form.Label>
                                <Form.Select
                                    value={formState.dateFormat}
                                    onChange={(e) => setFormState({...formState, dateFormat: e.target.value})}
                                >
                                    {dateFormats.map((df) => (
                                        <option key={df} value={df}>{df}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                        <Col md={6}>
                            <Form.Group controlId="formNumberingSystem">
                                <Form.Label>Numbering System</Form.Label>
                                <Form.Select
                                    value={formState.numberingSystem}
                                    onChange={(e) => setFormState({...formState, numberingSystem: e.target.value})}
                                >
                                    {(options?.numberingSystems ?? [settings.numberingSystem]).map((ns) => (
                                        <option key={ns} value={ns}>{ns}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                        </Col>
                    </Row>
                    <div style={{marginTop: '1rem'}}>
                        <Button variant="primary" onClick={handleSave} disabled={saving}>
                            {saving ? 'Saving...' : 'Save Settings'}
                        </Button>
                    </div>
                </Form>
            </div>
            <div style={{marginTop: '1rem'}}>
                <p>Download a backup of the database and uploaded files.</p>
                <Button onClick={handleDownloadBackup} disabled={backupLoading} variant="primary">
                    {backupLoading ? 'Preparing backup...' : 'Download Backup'}
                </Button>
            </div>
        </Container>
    );
};

export default SettingsPage;

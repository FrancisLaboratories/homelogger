import React, {useEffect, useState} from 'react';
import {Card, Button, Form, ListGroup} from 'react-bootstrap';
import TodoItem from './TodoItem';
import {SERVER_URL} from '@/pages/_app';

interface Props {
    applianceId?: number;
    spaceType?: string;
}

const TodosSection: React.FC<Props> = ({applianceId, spaceType}) => {
    const [todos, setTodos] = useState<Array<any>>([]);
    const [newLabel, setNewLabel] = useState('');
    const [sortOption, setSortOption] = useState<string>('created_desc');
    const [filterOption, setFilterOption] = useState<string>('all');

    const load = async () => {
        try {
            let url = `${SERVER_URL}/todo`;
            if (applianceId) url = `${url}?applianceId=${applianceId}`;
            else if (spaceType) url = `${url}?spaceType=${spaceType}`;

            const resp = await fetch(url);
            if (!resp.ok) return;
            const data = await resp.json();

            // Enrich todos with appliance names when applicable
            const applianceIds: number[] = Array.from(new Set(data.filter((t: any) => t.applianceId).map((t: any) => Number(t.applianceId))));
            const nameMap: Record<number, string> = {};
            await Promise.all(applianceIds.map(async (id) => {
                try {
                    const r = await fetch(`${SERVER_URL}/appliances/${id}`);
                    if (!r.ok) return;
                    const a = await r.json();
                    nameMap[id] = a.applianceName || `Appliance ${id}`;
                } catch (e) {
                    console.error('Error loading appliance name', e);
                }
            }));

            const enriched = data.map((t: any) => ({
                ...t,
                sourceLabel: t.applianceId ? nameMap[Number(t.applianceId)] : (t.spaceType || null),
            }));

            setTodos(enriched);
        } catch (err) {
            console.error('Error loading todos', err);
        }
    };

    useEffect(() => { load(); }, [applianceId, spaceType]);

    const handleAdd = async () => {
        if (!newLabel || newLabel.trim() === '') return;
        const body: any = { label: newLabel, checked: false, userid: '1' };
        if (applianceId) body.applianceId = applianceId;
        if (spaceType) body.spaceType = spaceType;

        try {
            const resp = await fetch(`${SERVER_URL}/todo/add`, {
                method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
            });
            if (!resp.ok) throw new Error('Add failed');
            const added = await resp.json();

            // Enrich added todo with source label if appliance
            if (added.applianceId) {
                try {
                    const r = await fetch(`${SERVER_URL}/appliances/${added.applianceId}`);
                    if (r.ok) {
                        const a = await r.json();
                        added.sourceLabel = a.applianceName || `Appliance ${added.applianceId}`;
                    }
                } catch (e) {
                    console.error('Error loading appliance name', e);
                }
            } else {
                added.sourceLabel = added.spaceType || null;
            }

            setTodos(prev => [...prev, added]);
            setNewLabel('');
        } catch (err) {
            console.error('Error adding todo', err);
        }
    };

    const handleDelete = (id: string) => {
        setTodos(prev => prev.filter(t => t.id !== id));
    };

    return (
        <Card>
            <Card.Body>
                <h5>To-dos</h5>
                <div style={{display: 'flex', gap: '8px', marginBottom: '8px'}}>
                    <Form.Select aria-label="Sort todos" value={sortOption} onChange={(e) => setSortOption(e.target.value)} style={{maxWidth: '220px'}}>
                        <option value="created_desc">Created (newest)</option>
                        <option value="created_asc">Created (oldest)</option>
                        <option value="label_asc">Label (A - Z)</option>
                        <option value="label_desc">Label (Z - A)</option>
                    </Form.Select>

                    <Form.Select aria-label="Filter todos" value={filterOption} onChange={(e) => setFilterOption(e.target.value)} style={{maxWidth: '180px'}}>
                        <option value="all">All</option>
                        <option value="completed">Completed</option>
                        <option value="not_completed">Not completed</option>
                    </Form.Select>
                </div>

                {todos.length === 0 ? (
                    <div>No to-dos</div>
                ) : (
                    <ListGroup>
                        {(() => {
                            const filtered = todos.filter((t: any) => {
                                if (filterOption === 'all') return true;
                                if (filterOption === 'completed') return !!t.checked;
                                if (filterOption === 'not_completed') return !t.checked;
                                return true;
                            });

                            const sorted = filtered.slice().sort((a: any, b: any) => {
                                const sa = (a.label || '').toString();
                                const sb = (b.label || '').toString();

                                const ca = a.createdAt || a.CreatedAt || a.created_at || null;
                                const cb = b.createdAt || b.CreatedAt || b.created_at || null;

                                if (sortOption === 'label_asc') return sa.localeCompare(sb);
                                if (sortOption === 'label_desc') return sb.localeCompare(sa);

                                const da = ca ? new Date(ca).getTime() : 0;
                                const db = cb ? new Date(cb).getTime() : 0;
                                if (sortOption === 'created_asc') return da - db || sa.localeCompare(sb);
                                return db - da || sa.localeCompare(sb);
                            });

                            return sorted.map((t: any) => (
                                <TodoItem key={t.id} id={t.id} label={t.label} checked={t.checked} onDelete={handleDelete} applianceId={t.applianceId} spaceType={t.spaceType} sourceLabel={t.sourceLabel} createdAt={t.createdAt || t.CreatedAt || t.created_at} />
                            ));
                        })()}
                    </ListGroup>
                )}

                <Form.Group controlId="todoAdd" style={{marginTop: '8px', display: 'flex'}}>
                    <Form.Control type="text" placeholder="New to-do" value={newLabel} onChange={(e) => setNewLabel(e.target.value)} />
                    <Button variant="primary" onClick={handleAdd} style={{marginLeft: '8px'}}>Add</Button>
                </Form.Group>
            </Card.Body>
        </Card>
    );
};

export default TodosSection;

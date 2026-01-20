import React, {useEffect, useState} from 'react';
import {Card, Button, Form, ListGroup, Modal} from 'react-bootstrap';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {SERVER_URL} from '@/pages/_app';

interface Props {
    applianceId?: number;
    spaceType?: string;
}

interface NoteItem {
    id: number;
    title: string;
    body: string;
    createdAt?: string;
}

const storageKey = (applianceId?: number, spaceType?: string) => {
    if (applianceId) return `homelogger_notes_appliance_${applianceId}`;
    if (spaceType) return `homelogger_notes_space_${spaceType}`;
    return `homelogger_notes_general`;
};

const NotesSection: React.FC<Props> = ({applianceId, spaceType}) => {
    const [notes, setNotes] = useState<NoteItem[]>([]);
    const [showModal, setShowModal] = useState(false);
    const [mode, setMode] = useState<'view' | 'edit' | 'add'>('view');
    const [viewing, setViewing] = useState<NoteItem | null>(null);
    const [editing, setEditing] = useState<NoteItem | null>(null);
    const [title, setTitle] = useState('');
    const [body, setBody] = useState('');
    const [preview, setPreview] = useState(false);

    useEffect(() => {
        const load = async () => {
            try {
                const params: string[] = [];
                if (applianceId) params.push(`applianceId=${applianceId}`);
                if (spaceType) params.push(`spaceType=${encodeURIComponent(String(spaceType))}`);
                const q = params.length > 0 ? `?${params.join('&')}` : '';
                const resp = await fetch(`${SERVER_URL}/notes${q}`);
                if (!resp.ok) {
                    console.error('Failed to load notes', resp.statusText);
                    return;
                }
                const data: NoteItem[] = await resp.json();
                setNotes(data || []);
            } catch (e) {
                console.error('Failed to load notes', e);
            }
        };
        load();
    }, [applianceId, spaceType]);

    const persist = (next: NoteItem[]) => {
        setNotes(next);
    };

    const handleAdd = () => {
        setEditing(null);
        setViewing(null);
        setTitle('');
        setBody('');
        setPreview(false);
        setMode('add');
        setShowModal(true);
    };

    const handleEdit = (n: NoteItem) => {
        setEditing(n);
        setViewing(n);
        setTitle(n.title);
        setBody(n.body);
        setPreview(false);
        setMode('edit');
        setShowModal(true);
    };

    const handleView = (n: NoteItem) => {
        (async () => {
            try {
                const resp = await fetch(`${SERVER_URL}/notes/${n.id}`);
                if (resp.ok) {
                    const full: NoteItem = await resp.json();
                    setViewing(full);
                } else {
                    setViewing(n);
                }
            } catch (e) {
                console.error('Error fetching note', e);
                setViewing(n);
            }
            setEditing(null);
            setMode('view');
            setShowModal(true);
        })();
    };

    const handleDelete = async (id: number | string) => {
        if (!window.confirm('Delete this note?')) return;
        try {
            const resp = await fetch(`${SERVER_URL}/notes/delete/${String(id)}`, { method: 'DELETE' });
            if (!resp.ok) throw new Error('Delete failed');
            persist(notes.filter(n => String(n.id) !== String(id)));
        } catch (e) {
            console.error('Error deleting note', e);
        }
    };

    const handleSave = () => {
        if (!title || title.trim() === '') {
            alert('Title is required');
            return;
        }

        (async () => {
            try {
                if (mode === 'edit' && editing) {
                    const resp = await fetch(`${SERVER_URL}/notes/update/${editing.id}`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ title, body }),
                    });
                    if (!resp.ok) throw new Error('Update failed');
                    const updated: NoteItem = await resp.json();
                    persist(notes.map(n => n.id === updated.id ? updated : n));
                } else {
                    // add
                    const payload: any = { title, body };
                    if (applianceId) payload.applianceId = applianceId;
                    if (spaceType) payload.spaceType = spaceType;
                    const resp = await fetch(`${SERVER_URL}/notes/add`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(payload),
                    });
                    if (!resp.ok) throw new Error('Add failed');
                    const added: NoteItem = await resp.json();
                    persist([...(notes || []), added]);
                }
            } catch (e) {
                console.error('Error saving note', e);
            } finally {
                setShowModal(false);
                setMode('view');
                setViewing(null);
                setEditing(null);
            }
        })();
    };

    return (
        <Card>
            <Card.Body>
                <h5>Notes</h5>
                {notes.length === 0 ? (
                    <div>No notes yet.</div>
                ) : (
                    <ListGroup>
                        {notes.slice().sort((a,b) => {
                            const ta = (a.createdAt || (a as any).CreatedAt || (a as any).created_at) ? new Date(a.createdAt || (a as any).CreatedAt || (a as any).created_at).getTime() : 0;
                            const tb = (b.createdAt || (b as any).CreatedAt || (b as any).created_at) ? new Date(b.createdAt || (b as any).CreatedAt || (b as any).created_at).getTime() : 0;
                            return tb - ta;
                        }).map(n => {
                            const created = n.createdAt || (n as any).CreatedAt || (n as any).created_at || '';
                            const createdLabel = created ? new Date(created).toLocaleString() : '';
                            return (
                                <ListGroup.Item key={n.id} action onClick={() => handleView(n)} style={{cursor: 'pointer'}}>
                                    <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
                                        <div style={{fontWeight: 600}}>{n.title}</div>
                                        <div style={{fontSize: '0.9em', color: '#666'}}>{createdLabel}</div>
                                    </div>
                                </ListGroup.Item>
                            );
                        })}
                    </ListGroup>
                )}

                <div style={{marginTop: '8px'}}>
                    <Button variant="primary" onClick={handleAdd}>Add Note</Button>
                </div>

                <Modal show={showModal} onHide={() => { setShowModal(false); setMode('view'); setViewing(null); setEditing(null); setTitle(''); setBody(''); setPreview(false); }} size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>{mode === 'add' ? 'Add Note' : mode === 'edit' ? 'Edit Note' : viewing ? viewing.title : 'Note'}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        {mode === 'view' && viewing && (
                            <div>
                                <div style={{marginBottom: '8px'}}>
                                    <ReactMarkdown remarkPlugins={[remarkGfm]}>{viewing.body || ''}</ReactMarkdown>
                                </div>
                            </div>
                        )}

                        {(mode === 'edit' || mode === 'add') && (
                            <Form>
                                <Form.Group controlId="noteTitle">
                                    <Form.Label>Title</Form.Label>
                                    <Form.Control type="text" value={title} onChange={(e) => setTitle(e.target.value)} />
                                </Form.Group>
                                <Form.Group controlId="noteBody" style={{marginTop: '8px'}}>
                                    <Form.Label>Body (Markdown supported)</Form.Label>
                                    <Form.Control as="textarea" rows={8} value={body} onChange={(e) => setBody(e.target.value)} />
                                </Form.Group>
                                <Form.Check type="checkbox" label="Preview" checked={preview} onChange={(e) => setPreview(e.target.checked)} style={{marginTop: '8px'}} />
                                {preview && (
                                    <div style={{border: '1px solid #ddd', padding: '8px', marginTop: '8px'}}>
                                        <ReactMarkdown remarkPlugins={[remarkGfm]}>{body || ''}</ReactMarkdown>
                                    </div>
                                )}
                            </Form>
                        )}
                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={() => { setShowModal(false); setMode('view'); setViewing(null); setEditing(null); setTitle(''); setBody(''); setPreview(false); }}>Close</Button>
                        {mode === 'view' && viewing && (
                            <>
                                <Button variant="secondary" onClick={() => handleEdit(viewing)}>Edit</Button>
                                <Button variant="danger" onClick={() => { handleDelete(viewing.id); setShowModal(false); setViewing(null); }}>Delete</Button>
                            </>
                        )}
                        {(mode === 'edit' || mode === 'add') && (
                            <Button variant="primary" onClick={handleSave}>{mode === 'edit' ? 'Save' : 'Add'}</Button>
                        )}
                    </Modal.Footer>
                </Modal>
            </Card.Body>
        </Card>
    );
};

export default NotesSection;

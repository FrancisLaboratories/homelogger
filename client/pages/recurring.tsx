import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Container, Form, Row, Table } from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import { SERVER_URL } from '@/pages/_app';

type BudgetCategory = {
  id: number;
  name: string;
};

type RecurringTask = {
  id: number;
  name: string;
  description: string;
  intervalValue: number;
  intervalUnit: string;
  nextDueDate: string;
  estimatedCost: number;
  referenceType: string;
  spaceType: string;
  applianceId?: number | null;
  categoryId?: number | null;
  autoCreateTodo: boolean;
  notes: string;
};

const RecurringPage: React.FC = () => {
  const [tasks, setTasks] = useState<RecurringTask[]>([]);
  const [categories, setCategories] = useState<BudgetCategory[]>([]);
  const [newTask, setNewTask] = useState({
    name: '',
    description: '',
    intervalValue: 1,
    intervalUnit: 'month',
    nextDueDate: '',
    estimatedCost: '',
    referenceType: '',
    spaceType: '',
    applianceId: '',
    categoryId: '',
    autoCreateTodo: false,
    notes: '',
  });

  const loadTasks = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/recurring`);
      if (!resp.ok) return;
      const data = await resp.json();
      setTasks(data);
    } catch (e) {
      console.error('Error loading recurring tasks', e);
    }
  };

  const loadCategories = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/budget/categories`);
      if (!resp.ok) return;
      const data = await resp.json();
      setCategories(data);
    } catch (e) {
      console.error('Error loading categories', e);
    }
  };

  useEffect(() => {
    loadTasks();
    loadCategories();
  }, []);

  const handleAddTask = async () => {
    if (!newTask.name) return;
    const resp = await fetch(`${SERVER_URL}/recurring/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newTask.name,
        description: newTask.description,
        intervalValue: Number(newTask.intervalValue),
        intervalUnit: newTask.intervalUnit,
        nextDueDate: newTask.nextDueDate,
        estimatedCost: Number(newTask.estimatedCost || 0),
        referenceType: newTask.referenceType,
        spaceType: newTask.spaceType,
        applianceId: newTask.applianceId ? Number(newTask.applianceId) : undefined,
        categoryId: newTask.categoryId ? Number(newTask.categoryId) : undefined,
        autoCreateTodo: !!newTask.autoCreateTodo,
        notes: newTask.notes,
      }),
    });
    if (!resp.ok) return;
    const created = await resp.json();
    setTasks((prev) => [...prev, created]);
    setNewTask({
      name: '',
      description: '',
      intervalValue: 1,
      intervalUnit: 'month',
      nextDueDate: '',
      estimatedCost: '',
      referenceType: '',
      spaceType: '',
      applianceId: '',
      categoryId: '',
      autoCreateTodo: false,
      notes: '',
    });
  };

  const handleDeleteTask = async (id: number) => {
    const resp = await fetch(`${SERVER_URL}/recurring/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) return;
    setTasks((prev) => prev.filter((t) => t.id !== id));
  };

  const categoryName = (id?: number | null) =>
    categories.find((c) => c.id === id)?.name || 'Uncategorized';

  return (
    <Container style={{ marginTop: '16px' }}>
      <MyNavbar />
      <h3>Recurring Maintenance</h3>

      <Row className="g-3">
        <Col lg={12}>
          <Card>
            <Card.Body>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Frequency</th>
                    <th>Next Due</th>
                    <th>Estimated</th>
                    <th>Category</th>
                    <th>Auto Todo</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {tasks.length === 0 ? (
                    <tr>
                      <td colSpan={7} style={{ textAlign: 'center' }}>No recurring tasks</td>
                    </tr>
                  ) : (
                    tasks.map((t) => (
                      <tr key={t.id}>
                        <td>{t.name}</td>
                        <td>{t.intervalValue} {t.intervalUnit}</td>
                        <td>{t.nextDueDate || '-'}</td>
                        <td>${Number(t.estimatedCost || 0).toFixed(2)}</td>
                        <td>{categoryName(t.categoryId)}</td>
                        <td>{t.autoCreateTodo ? 'Yes' : 'No'}</td>
                        <td style={{ width: 70 }}>
                          <Button variant="outline-danger" size="sm" onClick={() => handleDeleteTask(t.id)}>Delete</Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              <Form className="mt-3">
                <Row className="g-2">
                  <Col md={4}>
                    <Form.Control placeholder="Task name" value={newTask.name} onChange={(e) => setNewTask({ ...newTask, name: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Form.Control type="number" min={1} value={newTask.intervalValue} onChange={(e) => setNewTask({ ...newTask, intervalValue: Number(e.target.value) })} />
                  </Col>
                  <Col md={2}>
                    <Form.Select value={newTask.intervalUnit} onChange={(e) => setNewTask({ ...newTask, intervalUnit: e.target.value })}>
                      <option value="day">Day</option>
                      <option value="week">Week</option>
                      <option value="month">Month</option>
                      <option value="year">Year</option>
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Control type="date" value={newTask.nextDueDate} onChange={(e) => setNewTask({ ...newTask, nextDueDate: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Form.Control type="number" step="0.01" placeholder="Estimate" value={newTask.estimatedCost} onChange={(e) => setNewTask({ ...newTask, estimatedCost: e.target.value })} />
                  </Col>
                  <Col md={3}>
                    <Form.Select value={newTask.categoryId} onChange={(e) => setNewTask({ ...newTask, categoryId: e.target.value })}>
                      <option value="">Category</option>
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={3}>
                    <Form.Check type="checkbox" label="Auto-create todo" checked={newTask.autoCreateTodo} onChange={(e) => setNewTask({ ...newTask, autoCreateTodo: e.target.checked })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Description" value={newTask.description} onChange={(e) => setNewTask({ ...newTask, description: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Notes" value={newTask.notes} onChange={(e) => setNewTask({ ...newTask, notes: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddTask}>Add Task</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default RecurringPage;

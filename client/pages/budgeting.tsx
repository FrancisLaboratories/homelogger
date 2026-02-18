import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Container, Form, Row, Table } from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import { SERVER_URL } from '@/pages/_app';

type BudgetScenario = {
  id: number;
  name: string;
  startDate: string;
  horizonMonths: number;
  inflationRate: number;
  isActive: boolean;
  notes: string;
};

type BudgetCategory = {
  id: number;
  name: string;
  assetGroup: string;
  description: string;
  color: string;
};

type PlannedCost = {
  id: number;
  scenarioId?: number | null;
  categoryId?: number | null;
  sourceType: string;
  sourceId?: number | null;
  costDate: string;
  amount: number;
  notes: string;
};

type BudgetSummary = {
  scenario?: BudgetScenario | null;
  horizonMonths: number;
  totalPlanned: number;
  monthlySavings: number;
  upcoming30Days: number;
  upcoming90Days: number;
  plannedCostCount: number;
  categoryTotals: Record<string, number>;
  monthlyBuckets: Array<{ month: string; total: number }>;
};

const BudgetingPage: React.FC = () => {
  const [scenarios, setScenarios] = useState<BudgetScenario[]>([]);
  const [categories, setCategories] = useState<BudgetCategory[]>([]);
  const [plannedCosts, setPlannedCosts] = useState<PlannedCost[]>([]);
  const [selectedScenarioId, setSelectedScenarioId] = useState<number | null>(null);
  const [summary, setSummary] = useState<BudgetSummary | null>(null);

  const [newScenario, setNewScenario] = useState({
    name: '',
    startDate: '',
    horizonMonths: 12,
    inflationRate: 0,
    isActive: false,
    notes: '',
  });

  const [newCategory, setNewCategory] = useState({
    name: '',
    assetGroup: '',
    description: '',
    color: '',
  });

  const [newCost, setNewCost] = useState({
    scenarioId: '',
    categoryId: '',
    sourceType: 'upgrade',
    costDate: '',
    amount: '',
    notes: '',
  });

  const loadScenarios = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/budget/scenarios`);
      if (!resp.ok) return;
      const data = await resp.json();
      setScenarios(data);
      if (data.length > 0 && selectedScenarioId === null) {
        const active = data.find((s: BudgetScenario) => s.isActive) || data[0];
        setSelectedScenarioId(active.id);
      }
    } catch (e) {
      console.error('Error loading scenarios', e);
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

  const loadPlannedCosts = async (scenarioId: number | null) => {
    try {
      let url = `${SERVER_URL}/planned-costs`;
      if (scenarioId) {
        url = `${url}?scenarioId=${scenarioId}`;
      }
      const resp = await fetch(url);
      if (!resp.ok) return;
      const data = await resp.json();
      setPlannedCosts(data);
    } catch (e) {
      console.error('Error loading planned costs', e);
    }
  };

  useEffect(() => {
    loadScenarios();
    loadCategories();
  }, []);

  useEffect(() => {
    loadPlannedCosts(selectedScenarioId);
  }, [selectedScenarioId]);

  useEffect(() => {
    const loadSummary = async () => {
      try {
        let url = `${SERVER_URL}/budget/summary`;
        if (selectedScenarioId) {
          url = `${url}?scenarioId=${selectedScenarioId}`;
        }
        const resp = await fetch(url);
        if (!resp.ok) return;
        const data = await resp.json();
        setSummary(data);
      } catch (e) {
        console.error('Error loading budget summary', e);
      }
    };

    loadSummary();
  }, [selectedScenarioId]);

  const handleAddScenario = async () => {
    if (!newScenario.name) return;
    const resp = await fetch(`${SERVER_URL}/budget/scenarios/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newScenario.name,
        startDate: newScenario.startDate,
        horizonMonths: Number(newScenario.horizonMonths),
        inflationRate: Number(newScenario.inflationRate),
        isActive: !!newScenario.isActive,
        notes: newScenario.notes,
      }),
    });
    if (!resp.ok) return;
    const created = await resp.json();
    setScenarios((prev) => [...prev, created]);
    if (selectedScenarioId === null) setSelectedScenarioId(created.id);
    setNewScenario({ name: '', startDate: '', horizonMonths: 12, inflationRate: 0, isActive: false, notes: '' });
  };

  const handleAddCategory = async () => {
    if (!newCategory.name) return;
    const resp = await fetch(`${SERVER_URL}/budget/categories/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newCategory),
    });
    if (!resp.ok) return;
    const created = await resp.json();
    setCategories((prev) => [...prev, created]);
    setNewCategory({ name: '', assetGroup: '', description: '', color: '' });
  };

  const handleAddCost = async () => {
    if (!newCost.costDate || !newCost.amount) return;
    const resp = await fetch(`${SERVER_URL}/planned-costs/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        scenarioId: newCost.scenarioId ? Number(newCost.scenarioId) : undefined,
        categoryId: newCost.categoryId ? Number(newCost.categoryId) : undefined,
        sourceType: newCost.sourceType,
        costDate: newCost.costDate,
        amount: Number(newCost.amount),
        notes: newCost.notes,
      }),
    });
    if (!resp.ok) return;
    const created = await resp.json();
    setPlannedCosts((prev) => [...prev, created]);
    setNewCost({ scenarioId: '', categoryId: '', sourceType: 'upgrade', costDate: '', amount: '', notes: '' });
  };

  const handleDeleteCost = async (id: number) => {
    const resp = await fetch(`${SERVER_URL}/planned-costs/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) return;
    setPlannedCosts((prev) => prev.filter((c) => c.id !== id));
  };

  const categoryName = (id?: number | null) =>
    categories.find((c) => c.id === id)?.name || 'Uncategorized';

  return (
    <Container style={{ marginTop: '16px' }}>
      <MyNavbar />
      <h3>Budgeting</h3>

      <Row className="g-3" style={{ marginBottom: '12px' }}>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Monthly Savings Target</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.monthlySavings || 0).toFixed(2)}/mo</div>
              <div style={{ fontSize: '0.85rem' }}>{summary?.plannedCostCount || 0} items in plan</div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Upcoming (30 days)</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.upcoming30Days || 0).toFixed(2)}</div>
              <div style={{ fontSize: '0.85rem' }}>Next 90 days: ${(summary?.upcoming90Days || 0).toFixed(2)}</div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Total Planned</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.totalPlanned || 0).toFixed(2)}</div>
              <div style={{ fontSize: '0.85rem' }}>{summary?.horizonMonths || 0} month horizon</div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="g-3">
        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Scenarios</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Start</th>
                    <th>Horizon</th>
                    <th>Active</th>
                  </tr>
                </thead>
                <tbody>
                  {scenarios.length === 0 ? (
                    <tr>
                      <td colSpan={4} style={{ textAlign: 'center' }}>No scenarios</td>
                    </tr>
                  ) : (
                    scenarios.map((s) => (
                      <tr key={s.id} onClick={() => setSelectedScenarioId(s.id)} style={{ cursor: 'pointer' }}>
                        <td>{s.name}</td>
                        <td>{s.startDate || '-'}</td>
                        <td>{s.horizonMonths} mo</td>
                        <td>{s.isActive ? 'Yes' : 'No'}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              <Form className="mt-3">
                <Row className="g-2">
                  <Col md={6}>
                    <Form.Control placeholder="Scenario name" value={newScenario.name} onChange={(e) => setNewScenario({ ...newScenario, name: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control type="date" value={newScenario.startDate} onChange={(e) => setNewScenario({ ...newScenario, startDate: e.target.value })} />
                  </Col>
                  <Col md={4}>
                    <Form.Control type="number" min={1} placeholder="Horizon (months)" value={newScenario.horizonMonths} onChange={(e) => setNewScenario({ ...newScenario, horizonMonths: Number(e.target.value) })} />
                  </Col>
                  <Col md={4}>
                    <Form.Control type="number" step="0.1" placeholder="Inflation %" value={newScenario.inflationRate} onChange={(e) => setNewScenario({ ...newScenario, inflationRate: Number(e.target.value) })} />
                  </Col>
                  <Col md={4}>
                    <Form.Check type="checkbox" label="Active" checked={newScenario.isActive} onChange={(e) => setNewScenario({ ...newScenario, isActive: e.target.checked })} />
                  </Col>
                  <Col md={12}>
                    <Form.Control placeholder="Notes" value={newScenario.notes} onChange={(e) => setNewScenario({ ...newScenario, notes: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddScenario}>Add Scenario</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Categories</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Asset Group</th>
                    <th>Color</th>
                  </tr>
                </thead>
                <tbody>
                  {categories.length === 0 ? (
                    <tr>
                      <td colSpan={3} style={{ textAlign: 'center' }}>No categories</td>
                    </tr>
                  ) : (
                    categories.map((c) => (
                      <tr key={c.id}>
                        <td>{c.name}</td>
                        <td>{c.assetGroup || '-'}</td>
                        <td>{c.color || '-'}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              <Form className="mt-3">
                <Row className="g-2">
                  <Col md={6}>
                    <Form.Control placeholder="Category name" value={newCategory.name} onChange={(e) => setNewCategory({ ...newCategory, name: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Asset group" value={newCategory.assetGroup} onChange={(e) => setNewCategory({ ...newCategory, assetGroup: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Color" value={newCategory.color} onChange={(e) => setNewCategory({ ...newCategory, color: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Description" value={newCategory.description} onChange={(e) => setNewCategory({ ...newCategory, description: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddCategory}>Add Category</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Category Breakdown</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Category</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {!summary || Object.keys(summary.categoryTotals || {}).length === 0 ? (
                    <tr>
                      <td colSpan={2} style={{ textAlign: 'center' }}>No category data</td>
                    </tr>
                  ) : (
                    Object.entries(summary.categoryTotals)
                      .sort((a, b) => b[1] - a[1])
                      .map(([name, total]) => (
                        <tr key={name}>
                          <td>{name}</td>
                          <td>${Number(total || 0).toFixed(2)}</td>
                        </tr>
                      ))
                  )}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Monthly Buckets</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Month</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {!summary || (summary.monthlyBuckets || []).length === 0 ? (
                    <tr>
                      <td colSpan={2} style={{ textAlign: 'center' }}>No monthly data</td>
                    </tr>
                  ) : (
                    summary.monthlyBuckets.map((m) => (
                      <tr key={m.month}>
                        <td>{m.month}</td>
                        <td>${Number(m.total || 0).toFixed(2)}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={12}>
          <Card>
            <Card.Body>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h5>Planned Costs</h5>
                <Form.Select style={{ maxWidth: 260 }} value={selectedScenarioId ?? ''} onChange={(e) => setSelectedScenarioId(e.target.value ? Number(e.target.value) : null)}>
                  <option value="">All scenarios</option>
                  {scenarios.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </Form.Select>
              </div>
              <Table striped bordered hover size="sm" className="mt-2">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Amount</th>
                    <th>Category</th>
                    <th>Type</th>
                    <th>Notes</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {plannedCosts.length === 0 ? (
                    <tr>
                      <td colSpan={6} style={{ textAlign: 'center' }}>No planned costs</td>
                    </tr>
                  ) : (
                    plannedCosts.map((c) => (
                      <tr key={c.id}>
                        <td>{c.costDate || '-'}</td>
                        <td>${c.amount?.toFixed(2)}</td>
                        <td>{categoryName(c.categoryId)}</td>
                        <td>{c.sourceType || '-'}</td>
                        <td>{c.notes || '-'}</td>
                        <td style={{ width: 70 }}>
                          <Button variant="outline-danger" size="sm" onClick={() => handleDeleteCost(c.id)}>Delete</Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              <Form className="mt-3">
                <Row className="g-2">
                  <Col md={3}>
                    <Form.Select value={newCost.scenarioId} onChange={(e) => setNewCost({ ...newCost, scenarioId: e.target.value })}>
                      <option value="">Scenario</option>
                      {scenarios.map((s) => (
                        <option key={s.id} value={s.id}>{s.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={3}>
                    <Form.Select value={newCost.categoryId} onChange={(e) => setNewCost({ ...newCost, categoryId: e.target.value })}>
                      <option value="">Category</option>
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Select value={newCost.sourceType} onChange={(e) => setNewCost({ ...newCost, sourceType: e.target.value })}>
                      <option value="upgrade">Upgrade</option>
                      <option value="repair">Repair</option>
                      <option value="maintenance">Maintenance</option>
                      <option value="recurring">Recurring</option>
                      <option value="other">Other</option>
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Control type="date" value={newCost.costDate} onChange={(e) => setNewCost({ ...newCost, costDate: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Form.Control type="number" step="0.01" placeholder="Amount" value={newCost.amount} onChange={(e) => setNewCost({ ...newCost, amount: e.target.value })} />
                  </Col>
                  <Col md={10}>
                    <Form.Control placeholder="Notes" value={newCost.notes} onChange={(e) => setNewCost({ ...newCost, notes: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Button onClick={handleAddCost}>Add Cost</Button>
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

export default BudgetingPage;

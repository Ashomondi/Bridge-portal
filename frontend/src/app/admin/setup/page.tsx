'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Home, 
  Settings, 
  Users, 
  GraduationCap, 
  BookOpen, 
  Clock, 
  Calendar as CalendarIcon, 
  Bell, 
  Search, 
  ShieldCheck, 
  LogOut, 
  Loader2, 
  AlertCircle, 
  Check, 
  ArrowLeft, 
  ArrowRight, 
  HelpCircle, 
  X, 
  Plus 
} from 'lucide-react';

interface GradeStream {
  id: string;
  name: string; // 'East', 'West', etc.
}

interface GradeConfig {
  id: string;
  gradeLevel: string; // 'Grade 7 (Junior Secondary)', etc.
  capacity: number;
  isActive: boolean;
  streams: GradeStream[];
}

interface SetupWizardData {
  currentStep: number; // 1 to 6
  totalSteps: number;
  completionPercentage: number;
  autoSyncEnabled: boolean;
  stepsStatus: Array<{
    stepNumber: number;
    title: string;
    status: 'Done' | 'Active' | 'Upcoming' | 'Pending';
  }>;
  grades: GradeConfig[];
  checklist: Array<{
    id: string;
    task: string;
    isDone: boolean;
  }>;
}

export default function AdminSetupWizardPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('setup');

  // Dynamic state replacing hardcoded mocks
  const [setupData, setSetupData] = useState<SetupWizardData | null>(null);
  const [activeStep, setActiveStep] = useState(3); // Default to Step 3: Classes & Streams as shown in design
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // New stream input modal or inline state
  const [newStreamInput, setNewStreamInput] = useState<{ [key: string]: string }>({});

  // Fetch setup wizard data from backend API
  useEffect(() => {
    const fetchSetupWizard = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch(`/api/v1/admin/setup?step=${activeStep}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize setup wizard configuration from institutional registry.');
        }

        const data = await res.json();
        setSetupData(data);
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading wizard state.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchSetupWizard();
  }, [activeStep]);

  // Handle toggling grade active status
  const handleToggleGrade = async (gradeId: string, currentStatus: boolean) => {
    try {
      const res = await fetch(`/api/v1/admin/setup/grades/${gradeId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ isActive: !currentStatus }),
      });

      if (!res.ok) throw new Error('Failed to update grade status.');

      const updatedGrade = await res.json();
      if (setupData) {
        setSetupData({
          ...setupData,
          grades: setupData.grades.map(g => g.id === updatedGrade.id ? updatedGrade : g),
        });
      }
    } catch (err: any) {
      alert(err.message || 'Could not update grade.');
    }
  };

  // Handle adding a stream to a grade
  const handleAddStream = async (gradeId: string) => {
    const streamName = newStreamInput[gradeId];
    if (!streamName) return;

    try {
      const res = await fetch(`/api/v1/admin/setup/grades/${gradeId}/streams`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: streamName }),
      });

      if (!res.ok) throw new Error('Failed to add stream.');

      const updatedGrade = await res.json();
      if (setupData) {
        setSetupData({
          ...setupData,
          grades: setupData.grades.map(g => g.id === updatedGrade.id ? updatedGrade : g),
        });
        setNewStreamInput({ ...newStreamInput, [gradeId]: '' });
      }
    } catch (err: any) {
      alert(err.message || 'Could not add stream.');
    }
  };

  // Handle saving and proceeding to next step
  const handleContinue = async () => {
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/setup/progress', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ completedStep: activeStep }),
      });

      if (!res.ok) throw new Error('Failed to save step progress.');

      if (activeStep < 6) {
        setActiveStep(prev => prev + 1);
      } else {
        alert('Institutional onboarding setup completed successfully!');
        router.push('/admin');
      }
    } catch (err: any) {
      alert(err.message || 'Action failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#eaf4f3] via-[#f2f7f6] to-white flex text-slate-800">
      
      {/* Left Institutional Sidebar */}
      <aside className="w-64 bg-white border-r border-slate-200/90 flex flex-col justify-between p-5 shrink-0 hidden md:flex shadow-sm">
        <div className="space-y-6">
          <div>
            <h1 className="font-bold text-slate-900 text-lg tracking-tight">Ikonex Academy</h1>
            <p className="text-xs text-slate-500 font-medium">Lakeview Secondary</p>
          </div>

          <nav className="space-y-1 text-sm font-medium text-slate-600">
            <button onClick={() => router.push('/admin')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Home className="w-4 h-4 text-slate-400" /> Home
            </button>
            <button onClick={() => setActiveNav('setup')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <Settings className="w-4 h-4 text-[#008272]" /> Setup
            </button>
            <button onClick={() => router.push('/admin/staff')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Users className="w-4 h-4 text-slate-400" /> Staff
            </button>
            <button onClick={() => router.push('/admin/parents')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Users className="w-4 h-4 text-slate-400" /> Parents
            </button>
            <button onClick={() => router.push('/admin/teaching')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <BookOpen className="w-4 h-4 text-slate-400" /> Teaching
            </button>
            <button onClick={() => router.push('/admin/timetables')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Clock className="w-4 h-4 text-slate-400" /> Timetables
            </button>
            <button onClick={() => router.push('/admin/calendar')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <CalendarIcon className="w-4 h-4 text-slate-400" /> Calendar
            </button>
            <button onClick={() => router.push('/admin/students')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <GraduationCap className="w-4 h-4 text-slate-400" /> Students
            </button>
            <button onClick={() => router.push('/notifications')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Bell className="w-4 h-4 text-slate-400" /> Notifications
            </button>
          </nav>
        </div>

        <div className="pt-4 border-t border-slate-200 space-y-3">
          <div className="flex items-center gap-3 bg-slate-50 p-2.5 rounded-2xl border border-slate-200">
            <div className="w-8 h-8 rounded-xl bg-[#008272] text-white flex items-center justify-center text-xs font-bold">
              DA
            </div>
            <div className="leading-tight">
              <p className="text-xs font-bold text-slate-900">Ms. Achieng</p>
              <p className="text-[10px] text-slate-500 font-medium">Principal</p>
            </div>
          </div>
          <button onClick={() => router.push('/login')} className="w-full flex items-center gap-2 text-xs font-semibold text-rose-600 hover:text-rose-700 px-2 transition">
            <LogOut className="w-3.5 h-3.5" /> Sign out
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-w-0 overflow-y-auto">
        
        {/* Top Header Bar */}
        <div className="w-full bg-white/60 backdrop-blur border-b border-slate-200 px-6 py-2.5 flex items-center justify-between text-xs text-slate-500">
          <span className="px-3 py-1 bg-purple-50 text-purple-700 rounded-full font-medium border border-purple-200/60">
            Principal • Ms. Achieng
          </span>
          <span className="bg-white/80 px-3 py-1 rounded-full border border-slate-200 font-medium">
            Demo environment • sample data only
          </span>
          <div className="flex items-center gap-4">
            <button className="p-1.5 hover:bg-slate-100 rounded-full text-slate-600 transition"><Search className="w-4 h-4" /></button>
            <button className="p-1.5 hover:bg-slate-100 rounded-full text-slate-600 transition relative">
              <Bell className="w-4 h-4" />
              <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-rose-500"></span>
            </button>
            <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center text-xs font-bold">
              DA
            </div>
          </div>
        </div>

        <div className="p-6 md:p-10 space-y-6 max-w-7xl w-full mx-auto">
          
          {/* Top Wizard Branding Header Card */}
          <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-3.5">
              <div className="w-10 h-10 rounded-2xl bg-[#008272] text-white flex items-center justify-center font-bold shadow-md shadow-[#008272]/20">
                <Settings className="w-5 h-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="font-bold text-slate-900 text-lg">Ikonex Academy</h2>
                  <span className="px-2 py-0.5 bg-emerald-50 text-[#008272] font-bold text-[10px] rounded-md">Wizard Mode</span>
                </div>
                <p className="text-xs text-slate-500">Lakeview Secondary · Institutional Onboarding</p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-xs font-medium text-slate-600 bg-slate-50 px-3 py-1.5 rounded-full border border-slate-200 flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span> Auto-sync enabled
              </span>
              <button 
                onClick={() => router.push('/admin')}
                className="px-4 py-2 bg-white hover:bg-slate-50 border border-slate-200 font-semibold rounded-xl text-xs transition shadow-sm"
              >
                Save and exit
              </button>
            </div>
          </div>

          {/* Error Banner if API fails */}
          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {isLoading ? (
            <div className="bg-white border border-slate-200 rounded-[28px] py-48 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
              <p className="text-xs font-medium">Loading institutional wizard configuration...</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              
              {/* Left Column: Setup Progression Steps */}
              <div className="lg:col-span-3 space-y-4">
                <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-6">
                  <div>
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Setup Progression</span>
                    <div className="flex items-center justify-between mt-1">
                      <h3 className="font-extrabold text-slate-900 text-base">6 Core Steps</h3>
                      <span className="text-xs font-bold text-[#008272] bg-emerald-50 px-2 py-0.5 rounded-full">
                        {setupData?.completionPercentage || 50}% Complete
                      </span>
                    </div>
                  </div>

                  <div className="space-y-3 relative before:absolute before:left-4 before:top-3 before:bottom-3 before:w-0.5 before:bg-slate-100">
                    {setupData?.stepsStatus?.map((step) => {
                      const isCurrent = activeStep === step.stepNumber;
                      return (
                        <div 
                          key={step.stepNumber}
                          onClick={() => setActiveStep(step.stepNumber)}
                          className={`relative flex items-center gap-3 p-2.5 rounded-xl cursor-pointer transition ${
                            isCurrent ? 'bg-[#eaf4f3] border border-[#008272]/30 font-bold text-slate-900' : 'hover:bg-slate-50 text-slate-600'
                          }`}
                        >
                          <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold z-10 ${
                            step.status === 'Done' ? 'bg-[#008272] text-white' : 
                            isCurrent ? 'bg-[#008272] text-white ring-4 ring-[#eaf4f3]' : 'bg-slate-200 text-slate-600'
                          }`}>
                            {step.status === 'Done' ? <Check className="w-3.5 h-3.5" /> : step.stepNumber}
                          </div>
                          <div className="text-xs leading-tight">
                            <p className="font-bold">Step {step.stepNumber}</p>
                            <p className="text-[11px] text-slate-500">{step.title}</p>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  <div className="bg-[#f2f8f7] border border-slate-200 rounded-2xl p-4 space-y-1 text-xs">
                    <div className="flex items-center gap-1.5 font-bold text-[#008272]">
                      <ShieldCheck className="w-4 h-4" /> Kenya CBC Framework Compliant
                    </div>
                    <p className="text-[11px] text-slate-600 leading-relaxed">
                      Grades and assessment levels pre-calibrated to Ministry of Education guidelines.
                    </p>
                  </div>
                </div>
              </div>

              {/* Middle Column: Active Step Form (Step 3: Classes and Streams) */}
              <div className="lg:col-span-6 bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6">
                
                {/* Step Sub-tabs */}
                <div className="flex bg-slate-100 p-1 rounded-2xl border border-slate-200 text-xs font-semibold">
                  <button onClick={() => setActiveStep(3)} className={`flex-1 py-2 rounded-xl transition ${activeStep === 3 ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600'}`}>
                    Step 3 · Classes
                  </button>
                  <button onClick={() => setActiveStep(1)} className={`flex-1 py-2 rounded-xl transition ${activeStep === 1 ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600'}`}>
                    Step 1 · Profile
                  </button>
                  <button onClick={() => setActiveStep(2)} className={`flex-1 py-2 rounded-xl transition ${activeStep === 2 ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600'}`}>
                    Step 2 · Calendar
                  </button>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">STEP 3 OF 6</span>
                    <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-[#008272]">
                      3 Active Grades · 6 Configured Streams
                    </span>
                  </div>
                  <h3 className="text-2xl font-extrabold text-slate-900 tracking-tight">Classes and streams</h3>
                  <p className="text-xs text-slate-500">
                    Enable the grades taught at Lakeview Secondary and specify your streams.
                  </p>
                </div>

                {/* Grades Configuration List */}
                <div className="space-y-4">
                  {setupData?.grades?.map((grade) => (
                    <div key={grade.id} className={`border rounded-2xl p-4 space-y-3 transition ${grade.isActive ? 'bg-[#f8faf9] border-slate-200 shadow-sm' : 'bg-slate-50/50 border-slate-200/60 opacity-75'}`}>
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-bold text-slate-900 text-sm">{grade.gradeLevel}</h4>
                          <p className="text-[11px] text-slate-500">Standard intake capacity: {grade.capacity} students</p>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            checked={grade.isActive}
                            onChange={() => handleToggleGrade(grade.id, grade.isActive)}
                            className="sr-only peer"
                          />
                          <div className="w-11 h-6 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[#008272]"></div>
                        </label>
                      </div>

                      {grade.isActive && (
                        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-slate-200/60">
                          {grade.streams?.map((stream) => (
                            <span key={stream.id} className="inline-flex items-center gap-1.5 px-3 py-1 bg-white border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 shadow-sm">
                              Stream: {stream.name} <X className="w-3 h-3 text-slate-400 cursor-pointer hover:text-rose-600" />
                            </span>
                          ))}
                          
                          <div className="flex items-center gap-1">
                            <input 
                              type="text" 
                              placeholder="New stream..."
                              value={newStreamInput[grade.id] || ''}
                              onChange={(e) => setNewStreamInput({ ...newStreamInput, [grade.id]: e.target.value })}
                              className="bg-white border border-slate-200 rounded-xl px-2.5 py-1 text-xs w-28 focus:outline-none focus:border-[#008272]"
                            />
                            <button 
                              onClick={() => handleAddStream(grade.id)}
                              className="p-1.5 bg-[#008272] text-white rounded-xl hover:bg-[#006e61] transition"
                            >
                              <Plus className="w-3 h-3" />
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>

                {/* Footer Navigation Buttons */}
                <div className="flex items-center justify-between pt-4 border-t border-slate-100">
                  <button 
                    onClick={() => setActiveStep(prev => Math.max(1, prev - 1))}
                    className="px-4 py-2.5 border border-slate-200 rounded-xl font-semibold text-slate-600 hover:bg-slate-50 text-xs flex items-center gap-1.5"
                  >
                    <ArrowLeft className="w-3.5 h-3.5" /> Back
                  </button>
                  <div className="flex items-center gap-3">
                    <span className="text-[11px] text-emerald-700 bg-emerald-50 px-3 py-1 rounded-full font-semibold flex items-center gap-1">
                      <Check className="w-3 h-3" /> Saved
                    </span>
                    <button 
                      onClick={handleContinue}
                      disabled={isSubmitting}
                      className="px-6 py-2.5 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition flex items-center gap-2 shadow-md shadow-[#008272]/20"
                    >
                      {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                      Continue <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>

              </div>

              {/* Right Column: Tips & Quick Checklist */}
              <div className="lg:col-span-3 space-y-6">
                
                {/* Tips Card */}
                <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4 text-xs">
                  <div className="flex items-center gap-2 text-slate-900 font-bold text-sm">
                    <HelpCircle className="w-4 h-4 text-[#008272]" /> Tips for this step
                  </div>
                  <p className="text-slate-600 leading-relaxed">
                    You can add streams later when student numbers are confirmed.
                  </p>
                  <div className="space-y-2 pt-2 border-t border-slate-100 text-[11px] text-slate-500">
                    <p>💡 Parents see your phone and office hours when the assistant can't answer a question.</p>
                    <p>💡 Fee due dates follow the term start date.</p>
                  </div>
                </div>

                {/* Quick Checklist */}
                <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4 text-xs">
                  <div className="flex items-center justify-between">
                    <h3 className="font-bold text-slate-900 text-sm">Quick Checklist</h3>
                    <span className="font-bold text-[#008272] bg-emerald-50 px-2 py-0.5 rounded-full text-[10px]">3 / 4 Done</span>
                  </div>

                  <div className="space-y-2.5">
                    <div className="flex items-start gap-2.5 text-slate-700">
                      <Check className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                      <span>Registered institutional identity</span>
                    </div>
                    <div className="flex items-start gap-2.5 text-slate-700">
                      <Check className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                      <span>Academic term dates configured</span>
                    </div>
                    <div className="flex items-start gap-2.5 text-slate-700">
                      <Check className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                      <span>Junior secondary streams allocated</span>
                    </div>
                    <div className="flex items-start gap-2.5 text-slate-400">
                      <span className="w-4 h-4 rounded-full border border-slate-300 shrink-0 mt-0.5"></span>
                      <span>Hostel and dormitory rooms</span>
                    </div>
                  </div>
                </div>

                {/* Assistance Footer */}
                <div className="bg-[#f2f8f7] border border-slate-200/80 rounded-[28px] p-6 space-y-2 text-xs shadow-sm">
                  <h4 className="font-bold text-slate-900 flex items-center gap-1.5">
                    <HelpCircle className="w-4 h-4 text-[#008272]" /> Need assistance?
                  </h4>
                  <p className="text-slate-600 leading-relaxed">
                    Contact Ikonex Education Team
                  </p>
                </div>

              </div>

            </div>
          )}

        </div>

      </main>

    </div>
  );
}
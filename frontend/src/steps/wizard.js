import React, { useState } from 'react';
import Step1 from './step1.js';
import Step2 from './step2.js';
import Step3 from './step3.js';
import './wizard.css'; // Import the CSS file

const Wizard = () => {
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
  });

  const nextStep = () => {
    setStep(step + 1);
  };

  const prevStep = () => {
    setStep(step - 1);
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const submit = () => {
    console.log('Form submitted:', formData);
    // Add your form submission logic here
  };

  return (
    <div className="wizard-container">
      <div className="wizard">
        {step === 1 && (
          <Step1
            nextStep={nextStep}
            handleChange={handleChange}
            values={formData}
          />
        )}
        {step === 2 && (
          <Step2
            prevStep={prevStep}
            nextStep={nextStep}
            handleChange={handleChange}
            values={formData}
          />
        )}
        {step === 3 && (
          <Step3
            prevStep={prevStep}
            submit={submit}
            handleChange={handleChange}
            values={formData}
          />
        )}
      </div>
    </div>
  );
};

export default Wizard;
